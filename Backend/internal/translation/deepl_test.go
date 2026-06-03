package translation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeepLTranslatorTranslate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/translate" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "DeepL-Auth-Key test-key" {
			t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}

		var request deepLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if request.TargetLang != "UK" || request.SourceLang != "EN" {
			t.Fatalf("unexpected language pair: %+v", request)
		}
		if len(request.Text) != 2 || request.Text[0] != "Luxury" || request.Context == "" {
			t.Fatalf("unexpected request body: %+v", request)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"translations":[{"text":"Люкс"},{"text":"Позашляховик"}]}`))
	}))
	defer server.Close()

	translator := NewDeepLTranslator(" test-key ", server.URL+"/")
	translations, err := translator.Translate(context.Background(), "UK", []string{"Luxury", "SUV"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(translations) != 2 || translations[0] != "Люкс" || translations[1] != "Позашляховик" {
		t.Fatalf("unexpected translations: %+v", translations)
	}
}

func TestDeepLTranslatorUnavailableWithoutAPIKey(t *testing.T) {
	translator := NewDeepLTranslator(" ", "")
	_, err := translator.Translate(context.Background(), "UK", []string{"Luxury"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}

	var nilTranslator *DeepLTranslator
	_, err = nilTranslator.Translate(context.Background(), "UK", []string{"Luxury"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable for nil translator, got %v", err)
	}
}

func TestDeepLTranslatorHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	translator := NewDeepLTranslator("test-key", server.URL)
	_, err := translator.Translate(context.Background(), "UK", []string{"Luxury"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}

func TestDeepLTranslatorRejectsMismatchedTranslationCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"translations":[{"text":"Люкс"}]}`))
	}))
	defer server.Close()

	translator := NewDeepLTranslator("test-key", server.URL)
	_, err := translator.Translate(context.Background(), "UK", []string{"Luxury", "SUV"})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
}
