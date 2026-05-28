package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tattoo-consultation/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LoraService manages LoRA training and body-preview generation.
type LoraService struct {
	Pool         *pgxpool.Pool
	UploadDir    string
	ReplicateKey string
	S3           *storage.S3Client
}

// Lora represents a LoRA training record.
type Lora struct {
	ID                   string
	ConsultationID       string
	Status               string
	ReplicateTrainingID  string
	ReplicateVersionID   string
	LoraWeightsURL       string
	TriggerWord          string
	ErrorMessage         string
	TotalPhotos          int
}

// replicateTrainingInput is the request body for POST /trainings.
type replicateTrainingInput struct {
	Input struct {
		InputImages  string `json:"input_images"`   // ZIP URL
		TriggerWord  string `json:"trigger_word"`   // unique token
		Steps        int    `json:"steps"`          // 500-4000
		Resolution   string `json:"resolution"`     // "512,768,1024"
		Autocaption  bool   `json:"autocaption"`    // auto-caption with LLaVA
		LoraRank     int    `json:"lora_rank"`      // 16 default
	} `json:"input"`
}

// TrainLora kicks off LoRA training on Replicate.
// Returns the lora record (with DB ID) for polling.
func (s *LoraService) TrainLora(ctx context.Context, photoURLs []string, consultationID string) (*Lora, error) {
	// 1. Create lora record in DB
	var loraID string
	err := s.Pool.QueryRow(ctx,
		`INSERT INTO loras (consultation_id, status, total_photos)
		 VALUES ($1, 'uploading', $2) RETURNING id`,
		consultationID, len(photoURLs),
	).Scan(&loraID)
	if err != nil {
		return nil, fmt.Errorf("insert lora record: %w", err)
	}

	// 2. Download all photos and create ZIP
	zipPath := filepath.Join(s.UploadDir, "loras", loraID+".zip")
	if err := s.createZipFromURLs(photoURLs, zipPath); err != nil {
		s.failLora(ctx, loraID, "zip creation: "+err.Error())
		return nil, fmt.Errorf("create zip: %w", err)
	}

	// 3. Upload ZIP to S3 for Replicate to download
	zipKey := "loras/" + loraID + ".zip"
	zipURL := zipPath // local fallback
	if s.S3 != nil {
		s3URL, err := s.S3.UploadFile(ctx, zipPath, zipKey)
		if err != nil {
			s.failLora(ctx, loraID, "S3 upload ZIP: "+err.Error())
			return nil, fmt.Errorf("upload zip to s3: %w", err)
		}
		zipURL = s3URL
	}

	// 4. Generate trigger word
	triggerWord := "GVRBX" + loraID[:8]

	// 5. Kick off training on Replicate
	body := replicateTrainingInput{}
	body.Input.InputImages = zipURL
	body.Input.TriggerWord = triggerWord
	body.Input.Steps = 800 // fewer steps for body reference (vs full style training)
	body.Input.Resolution = "512,768,1024"
	body.Input.Autocaption = true
	body.Input.LoraRank = 16

	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		"https://api.replicate.com/v1/models/ostris/flux-dev-lora-trainer/versions/26dce37af90b9d997eeb970d92e47de3064d46c300504ae376c75bef6a9022d2/trainings",
		bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+s.ReplicateKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.failLora(ctx, loraID, "replicate API: "+err.Error())
		return nil, fmt.Errorf("replicate training API: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		s.failLora(ctx, loraID, fmt.Sprintf("replicate returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)])))
		return nil, fmt.Errorf("replicate training status %d", resp.StatusCode)
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		s.failLora(ctx, loraID, "unmarshal replicate: "+err.Error())
		return nil, fmt.Errorf("unmarshal training response: %w", err)
	}

	// 6. Update DB with training ID
	if _, err := s.Pool.Exec(ctx,
		`UPDATE loras SET status = 'training', replicate_training_id = $1, trigger_word = $2, updated_at = now() WHERE id = $3`,
		result.ID, triggerWord, loraID,
	); err != nil {
		return nil, fmt.Errorf("update lora training id: %w", err)
	}

	return &Lora{
		ID:                  loraID,
		ConsultationID:      consultationID,
		Status:              "training",
		ReplicateTrainingID: result.ID,
		TriggerWord:         triggerWord,
		TotalPhotos:         len(photoURLs),
	}, nil
}

// PollLora checks training status. Returns true if complete, false if still training.
func (s *LoraService) PollLora(ctx context.Context, lora *Lora) (bool, error) {
	req, _ := http.NewRequest("GET",
		"https://api.replicate.com/v1/trainings/"+lora.ReplicateTrainingID,
		nil)
	req.Header.Set("Authorization", "Bearer "+s.ReplicateKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("poll training: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Status  string `json:"status"`
		Error   string `json:"error"`
		Version string `json:"version"`
		Output  struct {
			Weights string `json:"weights"`
		} `json:"output"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return false, fmt.Errorf("unmarshal poll: %w", err)
	}

	switch result.Status {
	case "succeeded":
		weightsURL := result.Output.Weights
		if weightsURL == "" {
			s.failLora(ctx, lora.ID, "training succeeded but no weights URL")
			return true, fmt.Errorf("no weights URL")
		}
		if _, err := s.Pool.Exec(ctx,
			`UPDATE loras SET status = 'ready', replicate_version_id = $1, lora_weights_url = $2, updated_at = now() WHERE id = $3`,
			result.Version, weightsURL, lora.ID,
		); err != nil {
			return true, fmt.Errorf("update lora ready: %w", err)
		}
		lora.Status = "ready"
		lora.ReplicateVersionID = result.Version
		lora.LoraWeightsURL = weightsURL
		return true, nil

	case "failed", "canceled":
		s.failLora(ctx, lora.ID, result.Error)
		return true, fmt.Errorf("training failed: %s", result.Error)

	case "processing", "starting":
		return false, nil // still running

	default:
		return false, nil
	}
}

// WaitForLora blocks until training completes or timeout.
func (s *LoraService) WaitForLora(ctx context.Context, lora *Lora, timeout time.Duration) (*Lora, error) {
	deadline := time.After(timeout)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			s.failLora(ctx, lora.ID, "timeout waiting for LoRA training")
			return lora, fmt.Errorf("LoRA training timed out after %v", timeout)
		case <-ticker.C:
			done, err := s.PollLora(ctx, lora)
			if done {
				return lora, err // nil if succeeded, error if failed
			}
			// keep polling
		}
	}
}

// GenerateBodyPreview generates a body-superimposed tattoo preview using the trained LoRA.
// Uses img2img mode: the body photo as base, LoRA weights, and inpainting mask for the tattoo area.
func (s *LoraService) GenerateBodyPreview(ctx context.Context, lora *Lora, bodyPhotoURL, prompt, outputPath string) error {
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	body := map[string]interface{}{
		"input": map[string]interface{}{
			"prompt":             prompt + ", " + lora.TriggerWord + ", tattoo design on skin, photorealistic body tattoo preview, realistic skin texture",
			"image":              bodyPhotoURL,
			"replicate_weights":  lora.LoraWeightsURL,
			"lora_scale":         0.8,
			"num_inference_steps": 28,
			"guidance_scale":     3.5,
			"output_format":      "png",
			"output_quality":     95,
		},
	}

	payload, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		"https://api.replicate.com/v1/models/ostris/flux-dev-lora-trainer/predictions",
		bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+s.ReplicateKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "wait")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("body preview API: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("body preview returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
	}

	var result struct {
		Output string `json:"output"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("unmarshal body preview: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("body preview error: %s", result.Error)
	}
	if result.Output == "" {
		return fmt.Errorf("no output from body preview")
	}

	// Download the image
	imgResp, err := http.Get(result.Output)
	if err != nil {
		return fmt.Errorf("download body preview: %w", err)
	}
	defer imgResp.Body.Close()

	imgData, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return fmt.Errorf("read body preview: %w", err)
	}

	return os.WriteFile(outputPath, imgData, 0644)
}

// createZipFromURLs downloads all images from URLs and creates a ZIP archive.
func (s *LoraService) createZipFromURLs(urls []string, zipPath string) error {
	os.MkdirAll(filepath.Dir(zipPath), 0755)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for i, url := range urls {
		// Download image
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("download photo %d: %w", i, err)
		}
		imgData, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read photo %d: %w", i, err)
		}

		// Add to ZIP
		filename := fmt.Sprintf("photo_%03d.jpg", i+1)
		fw, err := zw.Create(filename)
		if err != nil {
			return fmt.Errorf("zip create entry %d: %w", i, err)
		}
		if _, err := fw.Write(imgData); err != nil {
			return fmt.Errorf("zip write entry %d: %w", i, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("zip close: %w", err)
	}

	return os.WriteFile(zipPath, buf.Bytes(), 0644)
}

// GetPhotosForConsultation returns all photo URLs for a consultation.
func (s *LoraService) GetPhotosForConsultation(ctx context.Context, consultationID string) ([]string, error) {
	rows, err := s.Pool.Query(ctx,
		`SELECT photo_url FROM consultation_photos WHERE consultation_id = $1 ORDER BY photo_order`,
		consultationID,
	)
	if err != nil {
		return nil, fmt.Errorf("query photos: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// failLora marks a LoRA training as failed.
func (s *LoraService) failLora(ctx context.Context, loraID, reason string) {
	s.Pool.Exec(ctx,
		`UPDATE loras SET status = 'failed', error_message = $1, updated_at = now() WHERE id = $2`,
		reason, loraID,
	)
}
