package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"autorent-backend/internal/models"
	"autorent-backend/internal/repositories"

	"google.golang.org/genai"
)

const searchAvailableCarsFunctionName = "search_available_cars"

var ErrAIUnavailable = errors.New("ai assistant is unavailable")

type CarRecommendationService struct {
	apiKey string
	model  string
	repo   repositories.CarRepository
}

func NewCarRecommendationService(apiKey string, model string, repo repositories.CarRepository) *CarRecommendationService {
	if strings.TrimSpace(model) == "" {
		model = "gemini-2.5-flash"
	}

	return &CarRecommendationService{
		apiKey: apiKey,
		model:  model,
		repo:   repo,
	}
}

func (s *CarRecommendationService) RecommendCar(ctx context.Context, message string) (*models.CarRecommendationResponse, error) {
	if strings.TrimSpace(s.apiKey) == "" {
		return nil, ErrAIUnavailable
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  s.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}

	config := s.generateContentConfig()

	firstResponse, err := client.Models.GenerateContent(
		ctx,
		s.model,
		genai.Text(s.buildUserPrompt(message)),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}

	functionCalls := firstResponse.FunctionCalls()
	if len(functionCalls) == 0 {
		return nil, fmt.Errorf("%w: gemini did not call %s", ErrAIUnavailable, searchAvailableCarsFunctionName)
	}

	functionCall := functionCalls[0]
	if functionCall.Name != searchAvailableCarsFunctionName {
		return nil, fmt.Errorf("%w: unexpected function call %s", ErrAIUnavailable, functionCall.Name)
	}

	filters := filtersFromFunctionArgs(functionCall.Args)
	log.Printf("Gemini function args: %+v", functionCall.Args)
	log.Printf("Parsed car filters: %+v", filters)

	cars, err := s.repo.FindAvailableCars(filters)
	if err != nil {
		return nil, err
	}
	if cars == nil {
		cars = make([]models.Car, 0)
	}
	log.Printf("Matched cars count: %d", len(cars))

	contents := []*genai.Content{
		genai.NewContentFromText(s.buildUserPrompt(message), genai.RoleUser),
		genai.NewContentFromFunctionCall(functionCall.Name, functionCall.Args, genai.RoleModel),
		genai.NewContentFromFunctionResponse(functionCall.Name, map[string]any{
			"cars":       cars,
			"filters":    filters,
			"currency":   "USD",
			"price_note": "All prices are in USD per day.",
		}, genai.RoleUser),
	}

	finalResponse, err := client.Models.GenerateContent(ctx, s.model, contents, config)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIUnavailable, err)
	}

	answer := strings.TrimSpace(finalResponse.Text())
	if answer == "" {
		answer = "I found matching cars, but could not generate a detailed explanation."
	}

	return &models.CarRecommendationResponse{
		Answer: answer,
		Cars:   cars,
	}, nil
}

func (s *CarRecommendationService) generateContentConfig() *genai.GenerateContentConfig {
	temperature := float32(0.2)

	return &genai.GenerateContentConfig{
		Temperature: &temperature,
		SystemInstruction: genai.NewContentFromText(
			"You are an AI car rental assistant for Autorent. "+
				"Understand user requests in English, Ukrainian, and Russian. "+
				"All rental prices in the car data are in USD per day. "+
				"Never convert prices to another currency unless the user explicitly asks for conversion. "+
				"Always write prices with the dollar sign, for example $220 per day. "+
				"Always use the search_available_cars function to find cars. "+
				"Do not invent cars. Recommend only cars returned by the function. "+
				"Preserve the order of cars returned by the function when listing recommendations. "+
				"If no cars are returned, politely explain that there are no matching cars. "+
				"Answer in the same language as the user's message.",
			genai.RoleUser,
		),
		Tools: []*genai.Tool{
			{
				FunctionDeclarations: []*genai.FunctionDeclaration{
					{
						Name:        searchAvailableCarsFunctionName,
						Description: "Search available rental cars by seats, daily budget, body type, fuel type, transmission, class, and usage purpose.",
						Parameters: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"min_seats": {
									Type:        genai.TypeInteger,
									Description: "Minimum number of seats required. Example: 5 for a family of five.",
								},
								"max_price_per_day": {
									Type:        genai.TypeNumber,
									Description: "Maximum daily rental price. The value must be interpreted as USD per day when the user mentions dollars, USD, $, долари, доларів, or долларов.",
								},
								"preferred_body_type": {
									Type:        genai.TypeString,
									Description: "Preferred body type, for example SUV, Sedan, Liftback.",
								},
								"preferred_fuel_type": {
									Type:        genai.TypeString,
									Description: "Preferred fuel type, for example Petrol, Diesel, Electric.",
								},
								"preferred_class": {
									Type:        genai.TypeString,
									Description: "Preferred car class, for example Premium, Luxury SUV, Electric Premium.",
								},
								"transmission": {
									Type:        genai.TypeString,
									Description: "Preferred transmission, for example Automatic.",
								},
								"purpose": {
									Type:        genai.TypeString,
									Description: "User purpose, for example family, business, travel, electric, luxury.",
								},
								"only_available": {
									Type:        genai.TypeBoolean,
									Description: "Whether to return only available cars. Usually true.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *CarRecommendationService) buildUserPrompt(message string) string {
	return fmt.Sprintf(
		"User request: %s\n\nExtract car rental requirements and call %s.",
		message,
		searchAvailableCarsFunctionName,
	)
}

func filtersFromFunctionArgs(args map[string]any) models.CarSearchFilters {
	return models.CarSearchFilters{
		MinSeats:          intFromAny(args["min_seats"]),
		MaxPricePerDay:    float64FromAny(args["max_price_per_day"]),
		PreferredBodyType: stringFromAny(args["preferred_body_type"]),
		PreferredFuelType: stringFromAny(args["preferred_fuel_type"]),
		PreferredClass:    stringFromAny(args["preferred_class"]),
		Transmission:      stringFromAny(args["transmission"]),
		Purpose:           stringFromAny(args["purpose"]),
		OnlyAvailable:     boolFromAny(args["only_available"], true),
	}
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func float64FromAny(value any) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(str)
}

func boolFromAny(value any, defaultValue bool) bool {
	if value == nil {
		return defaultValue
	}

	boolValue, ok := value.(bool)
	if !ok {
		return defaultValue
	}

	return boolValue
}
