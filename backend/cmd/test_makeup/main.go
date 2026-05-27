// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"tattoo-consultation/internal/service"
)

func main() {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENROUTER_API_KEY not set")
		os.Exit(1)
	}

	// Test AnalyzeExistingTattoo
	fmt.Println("=== Testing AnalyzeExistingTattoo ===")
	analysis, err := service.AnalyzeExistingTattoo(
		"/opt/tattoo-consultation/backend/uploads/b60f7815-d4e2-442d-a21d-9ccaffa73054.jpg",
		"Tôi muốn cover hình xăm này, làm mới lại màu sắc rực rỡ hơn",
		apiKey,
	)
	if err != nil {
		fmt.Printf("AnalyzeExistingTattoo FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Style: %s\n", analysis.TattooStyle)
	fmt.Printf("Lines: %s\n", analysis.LineQuality)
	fmt.Printf("Cover-up difficulty: %s\n", analysis.CoverUpDifficulty)
	fmt.Printf("Report: %.200s...\n", analysis.FullForensicReport)

	// Test CraftMakeupVariants
	fmt.Println("\n=== Testing CraftMakeupVariants ===")
	prompts, err := service.CraftMakeupVariants(
		"Tôi muốn cover hình xăm này, làm mới lại màu sắc rực rỡ hơn",
		analysis.FullForensicReport,
		apiKey,
	)
	if err != nil {
		fmt.Printf("CraftMakeupVariants FAILED: %v\n", err)
		os.Exit(1)
	}
	for i, p := range prompts {
		fmt.Printf("Variant %d: %s | %.100s...\n", i+1, p.Style, p.FullPrompt)
	}

	fmt.Println("\n=== ALL TESTS PASSED ===")
}

func init() {
	_ = context.Background
}
