package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CraftedPrompt holds a single design variant prompt.
type CraftedPrompt struct {
	Style       string `json:"style"`
	Composition string `json:"composition"`
	Mood        string `json:"mood"`
	FullPrompt  string `json:"full_prompt"`
}

// CraftMakeupVariants generates 3 tattoo improvement approaches (makeup/enhance/cover-up)
// based on forensic analysis of an existing tattoo.
func CraftMakeupVariants(clientConcern, forensicReport, apiKey string) ([]CraftedPrompt, error) {
	systemPrompt := `You are a world-class tattoo cover-up and enhancement specialist.
Given a forensic analysis of an EXISTING tattoo and the client's concern,
create 3 DISTINCT approaches to improve/transform it.

CRITICAL: The client ALREADY has a tattoo. Your job is NOT to design a new one from scratch
but to propose ways to WORK WITH or COVER the existing tattoo.

APPROACH 1 — "TOUCH-UP & REFRESH": Keep the existing design intact. Repair blown-out lines,
refresh faded colors, fix uneven shading, restore the original intent. Describe EXACTLY what
gets repaired and how. The final tattoo should look like a pristine, professional version of
the current one.

APPROACH 2 — "ENHANCE & EXTEND": Keep the existing tattoo as a foundation but add complementary
elements around/beyond it. Extend the composition, add background, incorporate the old work into
a larger piece. Describe what stays, what gets added, and how they integrate.

APPROACH 3 — "CREATIVE COVER-UP": Transform the existing tattoo into something new.
Use the old ink's shapes, shadows and lines as the foundation for a new design.
This is NOT "blast over" — the new design strategically incorporates the old one.
Describe the transformation strategy: what elements of the old tattoo become what elements
of the new design.

RULES:
- Each approach must explicitly reference specific elements of the existing tattoo
- Use the forensic analysis to inform what's fixable vs what needs covering
- Be HONEST about limitations: if a color is too dark to cover, say so
- Each prompt must be 300-400 words, English only, dense comma-separated paragraph
- Include: style, every element with hex colors, lighting, line weight, composition, texture
- End each prompt with: "tattoo flash art on white paper background, clean bold outlines, professional tattoo stencil, ultra-detailed, 8K resolution"

Output ONLY valid JSON array with 3 objects:
[{"style":"...","composition":"...","mood":"...","full_prompt":"..."}]`

	userPrompt := fmt.Sprintf(`Client's concern about their existing tattoo: %s

Forensic analysis of the existing tattoo: %s

Generate 3 approaches: Touch-up & Refresh, Enhance & Extend, Creative Cover-up.`,
		clientConcern, forensicReport)

	body := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  4096,
		"temperature": 0.7,
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
		return nil, fmt.Errorf("prompt API returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
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
		return nil, fmt.Errorf("no choices in prompt response")
	}

	content := extractJSON(result.Choices[0].Message.Content)

	var prompts []CraftedPrompt
	if err := json.Unmarshal([]byte(content), &prompts); err != nil {
		return nil, fmt.Errorf("unmarshal prompts: %w\nRaw: %s", err, content)
	}

	if len(prompts) != 3 {
		return nil, fmt.Errorf("expected 3 variants, got %d", len(prompts))
	}

	// Add the tattoo flash suffix to each prompt if not already present
	for i := range prompts {
		if !strings.Contains(prompts[i].FullPrompt, "tattoo flash art on white paper") {
			prompts[i].FullPrompt += ", tattoo flash art on white paper background, clean bold outlines, professional tattoo stencil, ultra-detailed, 8K resolution"
		}
	}

	return prompts, nil
}
func CraftVariants(ideaText, visionAnalysis, apiKey string) ([]CraftedPrompt, error) {
	systemPrompt := `You are a world-class tattoo art director. Given a client's idea and body analysis,
create 3 DISTINCT tattoo design prompts. Each must be different in style, composition, or mood.

RULES:
- Variant 1: FAITHFUL — exactly what the client asked for
- Variant 2: BOLD — same theme, more dynamic/aggressive composition
- Variant 3: ARTISTIC — same theme, different artistic style or color twist

Each prompt must be 300-400 words, English only, dense comma-separated paragraph.
Include: style, every element with hex colors, lighting, line weight, composition, texture, theme.
End each prompt with: "tattoo flash art on white paper background, clean bold outlines, professional tattoo stencil, ultra-detailed, 8K resolution"

Output ONLY valid JSON array with 3 objects:
[{"style":"...","composition":"...","mood":"...","full_prompt":"..."}]`

	userPrompt := fmt.Sprintf(`Client idea: %s

Body analysis from vision: %s

Generate 3 distinct tattoo design prompts.`, ideaText, visionAnalysis)

	body := map[string]interface{}{
		"model": "anthropic/claude-sonnet-4",
		"messages": []map[string]interface{}{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens":  4096,
		"temperature": 0.8,
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
		return nil, fmt.Errorf("prompt API returned %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
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
		return nil, fmt.Errorf("no choices in prompt response")
	}

	content := extractJSON(result.Choices[0].Message.Content)

	var prompts []CraftedPrompt
	if err := json.Unmarshal([]byte(content), &prompts); err != nil {
		return nil, fmt.Errorf("unmarshal prompts: %w\nRaw: %s", err, content)
	}

	if len(prompts) != 3 {
		return nil, fmt.Errorf("expected 3 variants, got %d", len(prompts))
	}

	// Add the tattoo flash suffix to each prompt if not already present
	for i := range prompts {
		if !strings.Contains(prompts[i].FullPrompt, "tattoo flash art on white paper") {
			prompts[i].FullPrompt += ", tattoo flash art on white paper background, clean bold outlines, professional tattoo stencil, ultra-detailed, 8K resolution"
		}
	}

	return prompts, nil
}
