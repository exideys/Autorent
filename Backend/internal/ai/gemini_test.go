package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	"google.golang.org/genai"
)

func TestUnavailableExtractor(t *testing.T) {
	_, err := UnavailableExtractor{}.ExtractCarFilters(context.Background(), "find a sedan")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

func TestNewGeminiExtractorRejectsEmptyAPIKey(t *testing.T) {
	extractor, err := NewGeminiExtractor(context.Background(), "  ", "")
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
	if extractor != nil {
		t.Fatalf("expected nil extractor, got %+v", extractor)
	}
}

func TestNewGeminiExtractorUsesDefaultModel(t *testing.T) {
	extractor, err := NewGeminiExtractor(context.Background(), "test-api-key", " ")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if extractor == nil || extractor.client == nil {
		t.Fatalf("expected configured extractor, got %+v", extractor)
	}
	if extractor.model != "gemini-2.5-flash" {
		t.Fatalf("expected default model, got %q", extractor.model)
	}
}

func TestGeminiConfigDefinesRequiredFunctionCall(t *testing.T) {
	cfg := geminiConfig()

	if cfg.Temperature == nil || *cfg.Temperature != 0 {
		t.Fatalf("expected zero temperature, got %+v", cfg.Temperature)
	}
	if cfg.MaxOutputTokens != 256 {
		t.Fatalf("unexpected max output tokens: %d", cfg.MaxOutputTokens)
	}
	if len(cfg.Tools) != 1 || len(cfg.Tools[0].FunctionDeclarations) != 1 {
		t.Fatalf("unexpected tools: %+v", cfg.Tools)
	}

	declaration := cfg.Tools[0].FunctionDeclarations[0]
	if declaration.Name != searchAvailableCarsFunction {
		t.Fatalf("unexpected function name: %q", declaration.Name)
	}
	if declaration.Parameters == nil || declaration.Parameters.Type != genai.TypeObject {
		t.Fatalf("unexpected schema: %+v", declaration.Parameters)
	}
	if _, ok := declaration.Parameters.Properties["preferred_brand"]; !ok {
		t.Fatalf("expected preferred_brand property: %+v", declaration.Parameters.Properties)
	}
	if cfg.ToolConfig == nil ||
		cfg.ToolConfig.FunctionCallingConfig == nil ||
		len(cfg.ToolConfig.FunctionCallingConfig.AllowedFunctionNames) != 1 ||
		cfg.ToolConfig.FunctionCallingConfig.AllowedFunctionNames[0] != searchAvailableCarsFunction {
		t.Fatalf("unexpected tool config: %+v", cfg.ToolConfig)
	}
}

func TestSystemInstructionMentionsSupportedLanguages(t *testing.T) {
	instruction := systemInstruction()
	if instruction == "" {
		t.Fatal("expected non-empty instruction")
	}
	for _, expected := range []string{"English", "Ukrainian", "function call"} {
		if !strings.Contains(instruction, expected) {
			t.Fatalf("instruction is missing %q: %q", expected, instruction)
		}
	}
}
