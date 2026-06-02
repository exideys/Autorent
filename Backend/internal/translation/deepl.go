package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const DefaultDeepLAPIURL = "https://api-free.deepl.com"

const deepLTranslationContext = "AutoRent car rental website UI. Translate short human-facing labels into natural Ukrainian, including vehicle classes, body types, fuel types, transmissions, statuses, filters, buttons, headings, form labels, helper text, and short descriptions. Vehicle class examples include Electric Premium, Luxury, Sport, Business, Economy, and Comfort. These are UI labels, not car brand or model names."

var ErrUnavailable = errors.New("translation service unavailable")

type DeepLTranslator struct {
	apiKey string
	apiURL string
	client *http.Client
}

type deepLRequest struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang"`
	Context    string   `json:"context,omitempty"`
}

type deepLResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

func NewDeepLTranslator(apiKey string, apiURL string) *DeepLTranslator {
	if strings.TrimSpace(apiURL) == "" {
		apiURL = DefaultDeepLAPIURL
	}

	return &DeepLTranslator{
		apiKey: strings.TrimSpace(apiKey),
		apiURL: strings.TrimRight(strings.TrimSpace(apiURL), "/"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *DeepLTranslator) Translate(ctx context.Context, targetLang string, texts []string) ([]string, error) {
	if t == nil || t.apiKey == "" {
		return nil, ErrUnavailable
	}

	requestBody, err := json.Marshal(deepLRequest{
		Text:       texts,
		TargetLang: targetLang,
		SourceLang: "EN",
		Context:    deepLTranslationContext,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal deepl request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, t.apiURL+"/v2/translate", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create deepl request: %w", err)
	}
	request.Header.Set("Authorization", "DeepL-Auth-Key "+t.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := t.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: deepl returned status %d", ErrUnavailable, response.StatusCode)
	}

	var payload deepLResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: decode deepl response: %v", ErrUnavailable, err)
	}
	if len(payload.Translations) != len(texts) {
		return nil, fmt.Errorf("%w: deepl returned %d translations for %d texts", ErrUnavailable, len(payload.Translations), len(texts))
	}

	translations := make([]string, 0, len(payload.Translations))
	for _, item := range payload.Translations {
		translations = append(translations, item.Text)
	}

	return translations, nil
}
