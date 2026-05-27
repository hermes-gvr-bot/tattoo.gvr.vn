package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// VisionAnalysis holds Claude Vision's analysis of a body photo.
type VisionAnalysis struct {
	BodyPart      string `json:"body_part"`
	SkinTone      string `json:"skin_tone"`
	SizeEstimates string `json:"size_estimates"`
	Lighting      string `json:"lighting"`
	FullAnalysis  string `json:"full_analysis"`
}

// AnalyzeBodyPhoto sends a body photo to Claude Vision for tattoo placement analysis.
func AnalyzeBodyPhoto(imagePath, apiKey string) (*VisionAnalysis, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(data)

	systemPrompt := `You are a world-class tattoo consultant analyzing a client's body photo.
Your task: determine WHERE on the body this photo shows, assess skin characteristics,
and estimate the suitable tattoo area. Output ONLY valid JSON with these keys:
body_part, skin_tone, size_estimates, lighting, full_analysis.
Be specific: "upper back between shoulder blades" not "back".
Estimate area in cm: "25cm wide x 15cm tall".`

	userPrompt := `Analyze this body photo for tattoo placement. Return JSON:
{
  "body_part": "specific anatomical location",
  "skin_tone": "fair / light / medium / olive / tan / dark — with undertones",
  "size_estimates": "available area estimate in cm (width x height)",
  "lighting": "natural / indoor / flash — note any shadows that affect analysis",
  "full_analysis": "detailed paragraph covering: muscle definition, existing tattoos, scars, moles, hair, skin texture, and overall suitability for tattooing"
}`

	body := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": userPrompt},
					{"type": "image_url", "image_url": map[string]string{
						"url": "data:image/jpeg;base64," + b64,
					}},
				},
			},
		},
		"max_tokens":   1024,
		"temperature":  0.3,
		"top_p":        1,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://tattoo.gvr.vn")
	req.Header.Set("X-Title", "Tattoo Consultation")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vision API returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := result.Choices[0].Message.Content

	// Extract JSON from response (Claude may wrap in ```json ... ```)
	content = extractJSON(content)

	var analysis VisionAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		// Fallback: store raw content as full_analysis
		analysis = VisionAnalysis{
			FullAnalysis: content,
		}
	}

	return &analysis, nil
}

// TattooForensicAnalysis holds deep forensic analysis of an existing tattoo.
type TattooForensicAnalysis struct {
	TattooStyle        string `json:"tattoo_style"`
	LineQuality        string `json:"line_quality"`
	ColorPalette       string `json:"color_palette"`
	Composition        string `json:"composition"`
	Condition          string `json:"condition"`
	ProblemAreas       string `json:"problem_areas"`
	CoverUpDifficulty  string `json:"cover_up_difficulty"`
	FullForensicReport string `json:"full_forensic_report"`
}

// AnalyzeExistingTattoo does deep forensic analysis of an existing tattoo
// for makeup/enhance/cover-up consultations.
func AnalyzeExistingTattoo(imagePath, clientConcern, apiKey string) (*TattooForensicAnalysis, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(data)

	systemPrompt := `You are a world-class tattoo forensic analyst and cover-up specialist with 20+ years of experience.
Your task: perform an EXTREMELY detailed forensic analysis of an EXISTING tattoo on a client's body.

Analyze EVERY aspect in microscopic detail — this will be used to plan touch-up, enhancement, or cover-up work.

CRITICAL ANALYSIS POINTS:
1. TATTOO STYLE: Identify exact style (e.g. Traditional American, Neo-Traditional, Japanese Irezumi, Blackwork,
   Realism, Tribal, Watercolor, Geometric, Minimalist, Old School, New School, Fine Line, etc.)
2. LINE QUALITY: Assess every line — are they crisp or blown out? Consistent weight? Any wobbles or breaks?
   Rate: Excellent / Good / Fair / Poor. Note WHERE the worst lines are.
3. COLOR PALETTE: List ALL colors present with their hex approximations. Note any fading or discoloration.
   Which colors held up best? Which faded worst? Are there any unwanted color shifts (e.g. black turning blue/green)?
4. COMPOSITION: How are elements arranged? Is there a focal point? Is the composition balanced?
   How does it flow with the body's anatomy?
5. CONDITION: Assess aging — how old does it look? Sun damage? Scarring within the tattoo?
   Ink spread (how much have lines thickened over time)? Any raised/keloid areas?
6. PROBLEM AREAS: List every specific problem — blowouts, uneven shading, patchy color, anatomical distortion,
   poor proportions, areas where the original artist made mistakes.
7. COVER-UP DIFFICULTY: Rate 1-10. Consider size, darkness, color saturation, scar tissue.
   Explain WHY this rating.
8. CLIENT CONCERN: The client says: "$CLIENT_CONCERN". Address this specifically in your report.

Output ONLY valid JSON with these keys:
tattoo_style, line_quality, color_palette, composition, condition,
problem_areas, cover_up_difficulty, full_forensic_report

The full_forensic_report should be a thorough 2-3 paragraph report covering all analysis above
in natural language, similar to what you'd write for another tattoo artist preparing for a cover-up job.`

	userPrompt := `Analyze this tattoo photo in extreme detail for a cover-up/enhancement consultation.
Client's concern: ` + clientConcern + `

Return JSON:
{
  "tattoo_style": "exact style identification with sub-style notes",
  "line_quality": "detailed line-by-line assessment with rating",
  "color_palette": "every color present, hex approximations, fading notes",
  "composition": "element arrangement, focal points, anatomical flow",
  "condition": "aging assessment, sun damage, ink spread, scarring",
  "problem_areas": "every specific flaw, blowout, mistake, distortion",
  "cover_up_difficulty": "rating 1-10 with detailed explanation",
  "full_forensic_report": "comprehensive 2-3 paragraph forensic report for tattoo artists"
}`

	body := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"type": "text", "text": userPrompt},
					{"type": "image_url", "image_url": map[string]string{
						"url": "data:image/jpeg;base64," + b64,
					}},
				},
			},
		},
		"max_tokens":  2048,
		"temperature": 0.3,
		"top_p":       1,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://tattoo.gvr.vn")
	req.Header.Set("X-Title", "Tattoo Consultation")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vision API returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := extractJSON(result.Choices[0].Message.Content)

	var analysis TattooForensicAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		// Fallback: store raw content as full report
		analysis = TattooForensicAnalysis{
			FullForensicReport: content,
		}
	}

	return &analysis, nil
}

// extractJSON strips markdown code fences and finds the JSON object.
func extractJSON(text string) string {
	text = strings.TrimSpace(text)
	// Strip ```json ... ``` or ``` ... ```
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}
	// Detect if it's a JSON array (starts with [)
	if strings.HasPrefix(text, "[") {
		return text
	}
	// Find the first { and last }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		extracted := text[start : end+1]
		// If the extracted text has multiple top-level objects (commas between } and {),
		// wrap it in an array
		if strings.Contains(extracted, "},{") || strings.Contains(extracted, "},\n{") || strings.Contains(extracted, "},\r\n{") {
			extracted = "[" + extracted + "]"
		}
		return extracted
	}
	return text
}
