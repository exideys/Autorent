package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"autorent-backend/internal/models"
	"autorent-backend/internal/repositories"

	"google.golang.org/genai"
)

const (
	searchAvailableCarsFunctionName = "search_available_cars"
	maxRecommendedCars              = 5
	maxGeminiRetries                = 2
)

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

	firstResponse, err := s.generateContentWithRetry(
		ctx,
		client,
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
	//log.Printf("Gemini function args: %+v", functionCall.Args)
	//log.Printf("Parsed car filters: %+v", filters)

	allCars, err := s.repo.FindAvailableCars(filters)
	if err != nil {
		return nil, err
	}

	if allCars == nil {
		allCars = make([]models.Car, 0)
	}

	totalMatches := len(allCars)
	recommendedCars := limitCars(allCars, maxRecommendedCars)

	log.Printf("Matched cars count: %d", totalMatches)
	log.Printf("Recommended cars returned to Gemini: %d", len(recommendedCars))

	//second gemini call for answer
	/*contents := []*genai.Content{
		genai.NewContentFromText(s.buildUserPrompt(message), genai.RoleUser),
		genai.NewContentFromFunctionCall(functionCall.Name, functionCall.Args, genai.RoleModel),
		genai.NewContentFromFunctionResponse(functionCall.Name, map[string]any{
			"recommended_cars": recommendedCars,
			"total_matches":    totalMatches,
			"filters":          filters,
			"currency":         "USD",
			"price_note":       "All prices are in USD per day.",
			"response_rule":    "total_matches means cars matching the hard constraints such as availability, seats, and budget. If the user provided soft preferences such as SUV, premium, business, comfort, or electric, say that total_matches cars match the basic constraints, then recommend the provided recommended_cars as the best preference-based options.",
		}, genai.RoleUser),
	}

	finalResponse, err := s.generateContentWithRetry(ctx, client, contents, config)
	if err != nil {
		log.Printf("Gemini final response failed, using fallback answer: %v", err)

		return &models.CarRecommendationResponse{
			Answer:       buildFallbackAnswer(message, recommendedCars, totalMatches),
			Cars:         recommendedCars,
			TotalMatches: totalMatches,
		}, nil
	}

	answer := strings.TrimSpace(finalResponse.Text())
	if answer == "" {
		answer = "I found matching cars, but could not generate a detailed explanation."
	}

	return &models.CarRecommendationResponse{
		Answer:       answer,
		Cars:         recommendedCars,
		TotalMatches: totalMatches,
	}, nil*/

	answer := buildBackendRecommendationAnswer(message, recommendedCars, totalMatches, filters)

	return &models.CarRecommendationResponse{
		Answer:       answer,
		Cars:         recommendedCars,
		TotalMatches: totalMatches,
	}, nil
}

func (s *CarRecommendationService) generateContentConfig() *genai.GenerateContentConfig {
	temperature := float32(0.2)

	return &genai.GenerateContentConfig{
		Temperature: &temperature,
		SystemInstruction: genai.NewContentFromText(
			"You are an AI car rental assistant for Autorent. "+
				"Understand user requests in English and Ukrainian. "+
				"All rental prices in the car data are in USD per day. "+
				"When the user asks for the cheapest, most expensive, newest, most powerful, or largest cars, set sort_by accordingly. "+
				"Never convert prices to another currency unless the user explicitly asks for conversion. "+
				"Always write prices with the dollar sign, for example $220 per day. "+
				"Always use the search_available_cars function to find cars. "+
				"Do not invent cars. Recommend only cars returned by the function. "+
				"Understand that total_matches is based on hard constraints only, while recommended_cars are the best ranked options based on user preferences. "+
				"Usually mention the total number of matching cars first. "+
				"Recommend only up to 5 best cars returned in recommended_cars. "+
				"Do not list every matching car if total_matches is greater than 5. "+
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
									Description: "Maximum daily rental price. The value must be interpreted as USD per day when the user mentions dollars, USD, $, долари, or доларів.",
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
								"sort_by": {
									Type:        genai.TypeString,
									Description: "Sorting preference. Use price_asc for cheapest cars, price_desc for most expensive cars, year_desc for newest cars, horsepower_desc for most powerful cars, seats_desc for largest cars. Use relevance when no explicit sorting is requested.",
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
		SortBy:            stringFromAny(args["sort_by"]),
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

func limitCars(cars []models.Car, limit int) []models.Car {
	if len(cars) <= limit {
		return cars
	}

	return cars[:limit]
}

func (s *CarRecommendationService) generateContentWithRetry(
	ctx context.Context,
	client *genai.Client,
	contents []*genai.Content,
	config *genai.GenerateContentConfig,
) (*genai.GenerateContentResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= maxGeminiRetries; attempt++ {
		response, err := client.Models.GenerateContent(ctx, s.model, contents, config)
		if err == nil {
			return response, nil
		}

		lastErr = err

		if isQuotaExceededError(err) {
			return nil, err
		}

		if attempt == maxGeminiRetries {
			break
		}

		log.Printf("Gemini request failed, retrying attempt %d/%d: %v", attempt+1, maxGeminiRetries, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
	}

	return nil, lastErr
}

func isQuotaExceededError(err error) bool {
	if err == nil {
		return false
	}

	errorText := strings.ToLower(err.Error())

	return strings.Contains(errorText, "429") ||
		strings.Contains(errorText, "resource_exhausted") ||
		strings.Contains(errorText, "quota exceeded")
}

func buildBackendRecommendationAnswer(message string, cars []models.Car, totalMatches int, filters models.CarSearchFilters) string {
	language := detectFallbackLanguage(message)

	if len(cars) == 0 {
		sortBy := strings.ToLower(strings.TrimSpace(filters.SortBy))
		switch language {
		case "uk":
			switch sortBy {
			case "price_asc":
				return fmt.Sprintf(
					"Ми знайшли %d доступних автомобілів. Нижче показано %d найдешевших варіантів.",
					totalMatches,
					len(cars),
				)
			case "price_desc":
				return fmt.Sprintf(
					"Ми знайшли %d доступних автомобілів. Нижче показано %d найдорожчих варіантів.",
					totalMatches,
					len(cars),
				)
			case "year_desc":
				return fmt.Sprintf(
					"Ми знайшли %d доступних автомобілів. Нижче показано %d найновіших варіантів.",
					totalMatches,
					len(cars),
				)
			case "horsepower_desc":
				return fmt.Sprintf(
					"Ми знайшли %d доступних автомобілів. Нижче показано %d найпотужніших варіантів.",
					totalMatches,
					len(cars),
				)
			default:
				return fmt.Sprintf(
					"Ми знайшли %d варіантів, які відповідають базовим критеріям. Нижче показано %d найкращих рекомендацій.",
					totalMatches,
					len(cars),
				)
			}

		default:
			switch sortBy {
			case "price_asc":
				return fmt.Sprintf(
					"We found %d available cars. Below are the %d cheapest options.",
					totalMatches,
					len(cars),
				)
			case "price_desc":
				return fmt.Sprintf(
					"We found %d available cars. Below are the %d most expensive options.",
					totalMatches,
					len(cars),
				)
			case "year_desc":
				return fmt.Sprintf(
					"We found %d available cars. Below are the %d newest options.",
					totalMatches,
					len(cars),
				)
			case "horsepower_desc":
				return fmt.Sprintf(
					"We found %d available cars. Below are the %d most powerful options.",
					totalMatches,
					len(cars),
				)
			default:
				return fmt.Sprintf(
					"We found %d options that match the basic criteria. Below are the top %d recommendations.",
					totalMatches,
					len(cars),
				)
			}
		}
	}

	switch language {
	case "uk":
		return fmt.Sprintf(
			"Ми знайшли %d варіантів, які відповідають базовим критеріям. Нижче показано %d найкращих рекомендацій.",
			totalMatches,
			len(cars),
		)
	default:
		return fmt.Sprintf(
			"We found %d options that match the basic criteria. Below are the top %d recommendations.",
			totalMatches,
			len(cars),
		)
	}
}

func detectFallbackLanguage(message string) string {
	lowerMessage := strings.ToLower(message)

	if strings.Contains(lowerMessage, "і") ||
		strings.Contains(lowerMessage, "ї") ||
		strings.Contains(lowerMessage, "є") ||
		strings.Contains(lowerMessage, "а") ||
		strings.Contains(lowerMessage, "мені") ||
		strings.Contains(lowerMessage, "потріб") ||
		strings.Contains(lowerMessage, "машин") ||
		strings.Contains(lowerMessage, "доларів") {
		return "uk"
	}

	return "en"
}
