package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tattoo-consultation/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Generator orchestrates the AI pipeline: vision → prompts → sketches → finals.
type Generator struct {
	Pool              *pgxpool.Pool
	UploadDir         string
	OpenRouterKey     string
	TogetherKey       string
	ReplicateKey      string
	S3                *storage.S3Client
	replicateLastCall time.Time
	replicateMu       sync.Mutex
}

// GenerateConsultation runs the full AI pipeline for a consultation.
// This should be called in a background goroutine.
func (g *Generator) GenerateConsultation(ctx context.Context, consultationID string) error {
	// 1. Load consultation data
	var userID, bodyPath, ideaText, consultationType, status string
	err := g.Pool.QueryRow(ctx,
		`SELECT user_id, body_photo_path, idea_text, status, COALESCE(consultation_type, 'new_tattoo') FROM consultations WHERE id = $1`,
		consultationID,
	).Scan(&userID, &bodyPath, &ideaText, &status, &consultationType)
	if err != nil {
		return fmt.Errorf("load consultation: %w", err)
	}

	// Update status to generating
	if _, err := g.Pool.Exec(ctx,
		`UPDATE consultations SET status = 'generating', updated_at = now() WHERE id = $1`,
		consultationID,
	); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// 2. Resolve body photo path (handle S3 URLs)
	fullPath := bodyPath
	if strings.HasPrefix(fullPath, "http://") || strings.HasPrefix(fullPath, "https://") {
		// Download from S3/HTTP to temp file
		tmpFile, err := os.CreateTemp("", "tattoo-body-*.jpg")
		if err != nil {
			g.failConsultation(ctx, consultationID, "temp file: "+err.Error())
			return fmt.Errorf("temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		resp, err := http.Get(fullPath)
		if err != nil {
			tmpFile.Close()
			g.failConsultation(ctx, consultationID, "download body photo: "+err.Error())
			return fmt.Errorf("download: %w", err)
		}
		io.Copy(tmpFile, resp.Body)
		resp.Body.Close()
		tmpFile.Close()
		fullPath = tmpFile.Name()
	} else if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(g.UploadDir, fullPath)
	}

	// 3. Analyze based on consultation type
	println("[generator]", consultationID, "type:", consultationType)
	if consultationType == "makeup_enhance" {
		println("[generator]", consultationID, "-> makeup_enhance path, calling AnalyzeExistingTattoo...")
		// FORENSIC ANALYSIS of existing tattoo
		tattooAnalysis, err := AnalyzeExistingTattoo(fullPath, ideaText, g.OpenRouterKey)
		if err != nil {
			println("[generator]", consultationID, "AnalyzeExistingTattoo FAILED:", err.Error())
			g.failConsultation(ctx, consultationID, "tattoo forensic analysis failed: "+err.Error())
			return fmt.Errorf("vision: %w", err)
		}

		// Save forensic analysis to DB
		analysisJSON, _ := json.Marshal(tattooAnalysis)
		// Truncate cover_up_difficulty to fit VARCHAR(100)
		diffShort := tattooAnalysis.CoverUpDifficulty
		if len(diffShort) > 100 {
			diffShort = diffShort[:97] + "..."
		}
		if _, err := g.Pool.Exec(ctx,
			`UPDATE consultations SET tattoo_analysis = $1, body_part = $2, skin_tone = $3, size_estimation = $4 WHERE id = $5`,
			string(analysisJSON), "existing_tattoo", "N/A (existing tattoo)", diffShort, consultationID,
		); err != nil {
			return fmt.Errorf("save analysis: %w", err)
		}

		// Craft 3 improvement approaches
		prompts, err := CraftMakeupVariants(ideaText, tattooAnalysis.FullForensicReport, g.OpenRouterKey)
		if err != nil {
			g.failConsultation(ctx, consultationID, "makeup prompt crafting failed: "+err.Error())
			return fmt.Errorf("prompts: %w", err)
		}

		// 4. Create variant records in DB
		variantIDs := make([]string, 3)
		for i, p := range prompts {
			var vid string
			err := g.Pool.QueryRow(ctx,
				`INSERT INTO variants (consultation_id, variant_number, prompt_used, sketch_status, final_status)
				 VALUES ($1, $2, $3, 'generating', 'pending') RETURNING id`,
				consultationID, i+1, p.FullPrompt,
			).Scan(&vid)
			if err != nil {
				return fmt.Errorf("insert variant %d: %w", i+1, err)
			}
			variantIDs[i] = vid
		}

		return g.generateVariants(ctx, consultationID, variantIDs, prompts)
	}

	// DEFAULT: New tattoo design flow
	analysis, err := AnalyzeBodyPhoto(fullPath, g.OpenRouterKey)
	if err != nil {
		g.failConsultation(ctx, consultationID, "vision analysis failed: "+err.Error())
		return fmt.Errorf("vision: %w", err)
	}

	// Save vision analysis to DB
	analysisJSON, _ := json.Marshal(analysis)
	if _, err := g.Pool.Exec(ctx,
		`UPDATE consultations SET vision_analysis = $1, body_part = $2, skin_tone = $3, size_estimation = $4 WHERE id = $5`,
		string(analysisJSON), analysis.BodyPart, analysis.SkinTone, analysis.SizeEstimates, consultationID,
	); err != nil {
		return fmt.Errorf("save analysis: %w", err)
	}

	// 3. Craft 3 variant prompts
	prompts, err := CraftVariants(ideaText, analysis.FullAnalysis, g.OpenRouterKey)
	if err != nil {
		g.failConsultation(ctx, consultationID, "prompt crafting failed: "+err.Error())
		return fmt.Errorf("prompts: %w", err)
	}

	// 4. Create variant records in DB
	variantIDs := make([]string, 3)
	for i, p := range prompts {
		var vid string
		err := g.Pool.QueryRow(ctx,
			`INSERT INTO variants (consultation_id, variant_number, prompt_used, sketch_status, final_status)
			 VALUES ($1, $2, $3, 'generating', 'pending') RETURNING id`,
			consultationID, i+1, p.FullPrompt,
		).Scan(&vid)
		if err != nil {
			return fmt.Errorf("insert variant %d: %w", i+1, err)
		}
		variantIDs[i] = vid
	}

	return g.generateVariants(ctx, consultationID, variantIDs, prompts)
}

// generateVariants generates all sketches and finals for the given variants.
// Shared between new_tattoo and makeup_enhance flows.
func (g *Generator) generateVariants(ctx context.Context, consultationID string, variantIDs []string, prompts []CraftedPrompt) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 6) // 3 sketches + 3 finals

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			vid := variantIDs[idx]
			prompt := prompts[idx].FullPrompt

			// Stagger to avoid API rate limits (Together + Replicate)
			time.Sleep(time.Duration(idx) * 15 * time.Second)

			// Sketch
			sketchKey := fmt.Sprintf("variants/%s_sketch.png", vid)
			sketchFullPath := filepath.Join(g.UploadDir, sketchKey)
			if err := g.generateSketch(prompt, sketchFullPath); err != nil {
				g.Pool.Exec(ctx, `UPDATE variants SET sketch_status = 'failed' WHERE id = $1`, vid)
				errCh <- fmt.Errorf("sketch variant %d: %w", idx+1, err)
				return
			}

			// Upload sketch to S3 if available
			sketchStorePath := "/uploads/" + sketchKey
			if g.S3 != nil {
				s3URL, err := g.S3.UploadFile(ctx, sketchFullPath, sketchKey)
				if err == nil {
					sketchStorePath = s3URL
				}
			}
			g.Pool.Exec(ctx,
				`UPDATE variants SET sketch_path = $1, sketch_status = 'completed' WHERE id = $2`,
				sketchStorePath, vid,
			)

			// Final
			g.Pool.Exec(ctx, `UPDATE variants SET final_status = 'generating' WHERE id = $1`, vid)
			finalKey := fmt.Sprintf("variants/%s_final.png", vid)
			finalFullPath := filepath.Join(g.UploadDir, finalKey)
			if err := g.generateFinal(prompt, finalFullPath); err != nil {
				g.Pool.Exec(ctx, `UPDATE variants SET final_status = 'failed' WHERE id = $1`, vid)
				errCh <- fmt.Errorf("final variant %d: %w", idx+1, err)
				return
			}

			// Upload final to S3 if available
			finalStorePath := "/uploads/" + finalKey
			if g.S3 != nil {
				s3URL, err := g.S3.UploadFile(ctx, finalFullPath, finalKey)
				if err == nil {
					finalStorePath = s3URL
				}
			}
			g.Pool.Exec(ctx,
				`UPDATE variants SET final_path = $1, final_status = 'completed' WHERE id = $2`,
				finalStorePath, vid,
			)
		}(i)
	}
	wg.Wait()
	close(errCh)

	// Collect errors
	var genErrors []error
	for e := range errCh {
		genErrors = append(genErrors, e)
	}

	if len(genErrors) > 0 {
		g.failConsultation(ctx, consultationID, fmt.Sprintf("%d generation errors: %v", len(genErrors), genErrors))
		return fmt.Errorf("generation errors: %v", genErrors)
	}

	// Mark consultation as completed
	if _, err := g.Pool.Exec(ctx,
		`UPDATE consultations SET status = 'completed', updated_at = now() WHERE id = $1`,
		consultationID,
	); err != nil {
		return fmt.Errorf("complete consultation: %w", err)
	}

	return nil
}

// generateSketch calls Together AI Flux Schnell (fast, cheap, $0.001).
func (g *Generator) generateSketch(prompt string, outputPath string) error {
	os.MkdirAll(filepath.Dir(outputPath), 0755)

	fullPrompt := prompt + ", tattoo flash art on white paper background, clean bold outlines, professional tattoo stencil, ultra-detailed"

	body := map[string]interface{}{
		"model":           "black-forest-labs/FLUX.1-schnell",
		"response_format": "b64_json",
		"prompt":          fullPrompt,
		"width":           1024,
		"height":          1024,
		"steps":           4,
	}

	// Retry loop with backoff for rate limiting
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt*attempt) * time.Second // 1s, 4s, 9s, 16s
			time.Sleep(delay)
		}

		payload, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "https://api.together.xyz/v1/images/generations", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+g.TogetherKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("together api: %w", err)
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 429 {
			lastErr = fmt.Errorf("together rate limited (attempt %d)", attempt+1)
			continue
		}
		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("together returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
			continue
		}

		var result struct {
			Data []struct {
				B64JSON string `json:"b64_json"`
			} `json:"data"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			lastErr = fmt.Errorf("unmarshal together: %w", err)
			continue
		}
		if len(result.Data) == 0 {
			lastErr = fmt.Errorf("no images from together")
			continue
		}

		imgData, err := base64.StdEncoding.DecodeString(result.Data[0].B64JSON)
		if err != nil {
			return fmt.Errorf("decode b64: %w", err)
		}

		return os.WriteFile(outputPath, imgData, 0644)
	}
	return fmt.Errorf("sketch failed after 5 attempts: %w", lastErr)
}

// generateFinal calls Replicate Flux 1.1 Pro (quality, $0.06).
func (g *Generator) generateFinal(prompt string, outputPath string) error {
	// Rate limit: Replicate free tier enforces 1 req/30s. Enforce 30s gap.
	g.replicateMu.Lock()
	elapsed := time.Since(g.replicateLastCall)
	if elapsed < 30*time.Second {
		time.Sleep(30*time.Second - elapsed)
	}
	g.replicateLastCall = time.Now()
	g.replicateMu.Unlock()

	os.MkdirAll(filepath.Dir(outputPath), 0755)

	fullPrompt := prompt + ", tattoo flash art on white paper, clean linework, 8K"

	body := map[string]interface{}{
		"input": map[string]interface{}{
			"prompt":            fullPrompt,
			"negative_prompt":   "blurry, photorealistic, 3D render, on skin, photograph, low quality",
			"width":             1024,
			"height":            1024,
			"num_inference_steps": 28,
			"guidance_scale":    3.5,
		},
	}

	// Retry loop with backoff (Replicate: 6 req/min, burst 1)
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt*5) * time.Second // 5s, 10s, 15s, 20s
			time.Sleep(delay)
		}

		payload, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST",
			"https://api.replicate.com/v1/models/black-forest-labs/flux-1.1-pro/predictions",
			bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+g.ReplicateKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "wait") // Synchronous

		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("replicate api: %w", err)
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == 429 {
			lastErr = fmt.Errorf("replicate rate limited (attempt %d)", attempt+1)
			continue
		}
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			lastErr = fmt.Errorf("replicate returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 300)]))
			continue
		}

		var result struct {
			Output string `json:"output"`
			Error  string `json:"error"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			lastErr = fmt.Errorf("unmarshal replicate: %w", err)
			continue
		}
		if result.Error != "" {
			lastErr = fmt.Errorf("replicate error: %s", result.Error)
			continue
		}
		if result.Output == "" {
			lastErr = fmt.Errorf("no output from replicate")
			continue
		}

		// Download the image
		imgResp, err := http.Get(result.Output)
		if err != nil {
			lastErr = fmt.Errorf("download final: %w", err)
			continue
		}
		defer imgResp.Body.Close()

		imgData, err := io.ReadAll(imgResp.Body)
		if err != nil {
			lastErr = fmt.Errorf("read final: %w", err)
			continue
		}

		return os.WriteFile(outputPath, imgData, 0644)
	}
	return fmt.Errorf("final failed after 5 attempts: %w", lastErr)
}

// failConsultation marks consultation as failed with error note.
func (g *Generator) failConsultation(ctx context.Context, consultationID, reason string) {
	g.Pool.Exec(ctx,
		`UPDATE consultations SET status = 'rejected', admin_notes = $1, updated_at = now() WHERE id = $2`,
		reason, consultationID,
	)
}
