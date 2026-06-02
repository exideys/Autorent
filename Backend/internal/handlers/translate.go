package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	translationUnavailableMessage = "Translation service is temporarily unavailable. Please try again later."
	invalidTranslationMessage     = "Invalid translation request."
	maxTranslationTexts           = 50
)

type Translator interface {
	Translate(ctx context.Context, targetLang string, texts []string) ([]string, error)
}

type TranslationHandler struct {
	translator Translator
}

type translationRequest struct {
	TargetLang string   `json:"target_lang"`
	Texts      []string `json:"texts"`
}

type translationResponse struct {
	Translations []string `json:"translations"`
}

func NewTranslationHandler(translator Translator) *TranslationHandler {
	return &TranslationHandler{translator: translator}
}

func RegisterTranslationRoutes(router gin.IRouter, translator Translator) {
	handler := NewTranslationHandler(translator)

	router.POST("/translate", handler.Translate)
}

func (h *TranslationHandler) Translate(c *gin.Context) {
	var input translationRequest
	if err := c.ShouldBindJSON(&input); err != nil || !validTranslationRequest(input) {
		respondError(c, http.StatusBadRequest, invalidTranslationMessage)
		return
	}

	if h.translator == nil {
		respondError(c, http.StatusServiceUnavailable, translationUnavailableMessage)
		return
	}

	translations, err := h.translator.Translate(c.Request.Context(), input.TargetLang, input.Texts)
	if err != nil {
		log.Printf("translation failed: %v", err)
		respondError(c, http.StatusServiceUnavailable, translationUnavailableMessage)
		return
	}

	if len(translations) != len(input.Texts) {
		log.Printf("translation failed: expected %d translations, got %d", len(input.Texts), len(translations))
		respondError(c, http.StatusServiceUnavailable, translationUnavailableMessage)
		return
	}

	c.JSON(http.StatusOK, translationResponse{Translations: translations})
}

func validTranslationRequest(input translationRequest) bool {
	if input.TargetLang != "UK" {
		return false
	}
	if len(input.Texts) == 0 || len(input.Texts) > maxTranslationTexts {
		return false
	}

	for _, text := range input.Texts {
		if strings.TrimSpace(text) == "" {
			return false
		}
	}

	return true
}
