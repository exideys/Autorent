package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"autorent-backend/internal/models"
	"autorent-backend/internal/recommendation"

	"google.golang.org/genai"
)

const searchAvailableCarsFunction = "search_available_cars"

var ErrUnavailable = errors.New("ai car assistant unavailable")

type CarFilterExtractor interface {
	ExtractCarFilters(ctx context.Context, message string) (models.CarRecommendationFilters, error)
}

type UnavailableExtractor struct{}

func (UnavailableExtractor) ExtractCarFilters(context.Context, string) (models.CarRecommendationFilters, error) {
	return models.CarRecommendationFilters{}, ErrUnavailable
}

type GeminiExtractor struct {
	client *genai.Client
	model  string
}

func NewGeminiExtractor(ctx context.Context, apiKey string, model string) (*GeminiExtractor, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrUnavailable
	}
	if strings.TrimSpace(model) == "" {
		model = "gemini-2.5-flash"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	return &GeminiExtractor{
		client: client,
		model:  model,
	}, nil
}

func (e *GeminiExtractor) ExtractCarFilters(ctx context.Context, message string) (models.CarRecommendationFilters, error) {
	response, err := e.client.Models.GenerateContent(
		ctx,
		e.model,
		genai.Text(message),
		geminiConfig(),
	)
	if err != nil {
		return models.CarRecommendationFilters{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}

	for _, call := range response.FunctionCalls() {
		if call.Name != searchAvailableCarsFunction {
			continue
		}

		filters := models.CarRecommendationFilters{OnlyAvailable: true}
		if len(call.Args) == 0 {
			return recommendation.NormalizeFilters(filters), nil
		}

		args, err := json.Marshal(call.Args)
		if err != nil {
			return models.CarRecommendationFilters{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}
		if err := json.Unmarshal(args, &filters); err != nil {
			return models.CarRecommendationFilters{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
		}

		return recommendation.NormalizeFilters(filters), nil
	}

	return models.CarRecommendationFilters{}, ErrUnavailable
}

func geminiConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemInstruction(), genai.RoleUser),
		Temperature:       genai.Ptr[float32](0),
		MaxOutputTokens:   256,
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        searchAvailableCarsFunction,
						Description: "Extract car rental search filters from a user request so the backend can search available cars in the database.",
						Parameters:  searchAvailableCarsSchema(),
					},
				},
			},
		},
		ToolConfig: &genai.ToolConfig{
			FunctionCallingConfig: &genai.FunctionCallingConfig{
				Mode:                 genai.FunctionCallingConfigModeAny,
				AllowedFunctionNames: []string{searchAvailableCarsFunction},
			},
		},
	}
}

func systemInstruction() string {
	return strings.Join([]string{
		"Understand user requests in English and Ukrainian.",
		"All prices in the database are USD per day.",
		"Do not invent cars.",
		"Always use the function call to extract filters.",
		"If the user mentions a specific brand or manufacturer, set preferred_brand.",
		"If the user asks for cheapest, most expensive, newest, most powerful, or largest cars, set sort_by.",
		"Do not convert currencies unless the user explicitly asks for conversion.",
	}, " ")
}

func searchAvailableCarsSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"min_seats": {
				Type:        genai.TypeInteger,
				Description: "Minimum required seats.",
			},
			"max_price_per_day": {
				Type:        genai.TypeNumber,
				Description: "Maximum daily rental price in USD.",
			},
			"preferred_brand": {
				Type:        genai.TypeString,
				Description: "Specific brand requested by the user, for example Tesla, BMW, Audi, Mercedes-Benz, Lexus, Porsche.",
			},
			"preferred_body_type": {
				Type:        genai.TypeString,
				Description: "Preferred body type such as SUV, Sedan, Hatchback, Liftback, Minivan, or Van.",
			},
			"preferred_fuel_type": {
				Type:        genai.TypeString,
				Description: "Preferred fuel type: Electric, Petrol, Diesel, or Hybrid.",
			},
			"preferred_class": {
				Type:        genai.TypeString,
				Description: "Preferred class such as Premium, Luxury, Economy, or Electric Comfort.",
			},
			"transmission": {
				Type:        genai.TypeString,
				Description: "Transmission preference: Automatic or Manual.",
			},
			"purpose": {
				Type:        genai.TypeString,
				Description: "Rental purpose such as family, business, comfort, travel, or electric.",
			},
			"sort_by": {
				Type:        genai.TypeString,
				Description: "Use price_asc for cheapest cars, price_desc for most expensive cars, year_desc for newest cars, horsepower_desc for most powerful cars, seats_desc for largest cars, and relevance or empty when no explicit sorting is requested.",
			},
			"only_available": {
				Type:        genai.TypeBoolean,
				Description: "Whether the search should include only cars with status available. Usually true.",
			},
		},
	}
}
