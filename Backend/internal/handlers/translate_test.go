package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeTranslator struct {
	translations []string
	err          error
	called       bool
	targetLang   string
	texts        []string
}

func (f *fakeTranslator) Translate(_ context.Context, targetLang string, texts []string) ([]string, error) {
	f.called = true
	f.targetLang = targetLang
	f.texts = append([]string(nil), texts...)
	if f.err != nil {
		return nil, f.err
	}
	return append([]string(nil), f.translations...), nil
}

func TestTranslateRejectsEmptyRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	translator := &fakeTranslator{}
	router := newTranslationTestRouter(translator)

	recorder := performTranslationRequest(t, router, `{}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if translator.called {
		t.Fatal("translator should not be called for invalid request")
	}
	if !strings.Contains(recorder.Body.String(), invalidTranslationMessage) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestTranslateRejectsUnsupportedTargetLang(t *testing.T) {
	gin.SetMode(gin.TestMode)

	translator := &fakeTranslator{}
	router := newTranslationTestRouter(translator)

	recorder := performTranslationRequest(t, router, `{"target_lang":"DE","texts":["Find cars"]}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if translator.called {
		t.Fatal("translator should not be called for unsupported target language")
	}
}

func TestTranslateRejectsEmptyStrings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	translator := &fakeTranslator{}
	router := newTranslationTestRouter(translator)

	recorder := performTranslationRequest(t, router, `{"target_lang":"UK","texts":["Find cars","   "]}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if translator.called {
		t.Fatal("translator should not be called for empty strings")
	}
}

func TestTranslateReturnsUnavailableWhenDeepLAPIKeyMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTranslationTestRouter(nil)

	recorder := performTranslationRequest(t, router, `{"target_lang":"UK","texts":["Find cars"]}`)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), translationUnavailableMessage) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestTranslateReturnsUnavailableWhenTranslatorFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := newTranslationTestRouter(&fakeTranslator{err: errors.New("deepl unavailable")})

	recorder := performTranslationRequest(t, router, `{"target_lang":"UK","texts":["Find cars"]}`)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}

func TestTranslatePreservesTranslationOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	translator := &fakeTranslator{translations: []string{"first translated", "second translated", "third translated"}}
	router := newTranslationTestRouter(translator)

	recorder := performTranslationRequest(t, router, `{"target_lang":"UK","texts":["First","Second","Third"]}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if !translator.called {
		t.Fatal("expected translator to be called")
	}
	if translator.targetLang != "UK" {
		t.Fatalf("expected target lang UK, got %q", translator.targetLang)
	}
	if strings.Join(translator.texts, ",") != "First,Second,Third" {
		t.Fatalf("unexpected translated texts: %#v", translator.texts)
	}

	var response translationResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if strings.Join(response.Translations, ",") != "first translated,second translated,third translated" {
		t.Fatalf("unexpected translations: %#v", response.Translations)
	}
}

func newTranslationTestRouter(translator Translator) *gin.Engine {
	router := gin.New()
	RegisterTranslationRoutes(router.Group("/api"), translator)
	return router
}

func performTranslationRequest(t *testing.T, router *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, "/api/translate", strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
