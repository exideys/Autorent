package recommendation

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	"autorent-backend/internal/models"
)

const maxRecommendations = 5

func NormalizeFilters(filters models.CarRecommendationFilters) models.CarRecommendationFilters {
	filters.PreferredBrand = strings.TrimSpace(filters.PreferredBrand)
	filters.PreferredBodyType = strings.TrimSpace(filters.PreferredBodyType)
	filters.PreferredFuelType = strings.TrimSpace(filters.PreferredFuelType)
	filters.PreferredClass = strings.TrimSpace(filters.PreferredClass)
	filters.Transmission = strings.TrimSpace(filters.Transmission)
	filters.Purpose = strings.TrimSpace(filters.Purpose)
	filters.SortBy = normalizeSort(filters.SortBy)

	if filters.MinSeats < 0 {
		filters.MinSeats = 0
	}
	if filters.MaxPricePerDay < 0 {
		filters.MaxPricePerDay = 0
	}

	return filters
}

func TopCars(cars []models.Car, filters models.CarRecommendationFilters) []models.Car {
	ranked := RankCars(cars, filters)
	if len(ranked) > maxRecommendations {
		return ranked[:maxRecommendations]
	}
	return ranked
}

func RankCars(cars []models.Car, filters models.CarRecommendationFilters) []models.Car {
	filters = NormalizeFilters(filters)
	ranked := append([]models.Car(nil), cars...)

	sort.SliceStable(ranked, func(i, j int) bool {
		first := ranked[i]
		second := ranked[j]

		switch filters.SortBy {
		case "price_asc":
			if first.PricePerDay != second.PricePerDay {
				return first.PricePerDay < second.PricePerDay
			}
		case "price_desc":
			if first.PricePerDay != second.PricePerDay {
				return first.PricePerDay > second.PricePerDay
			}
		case "year_desc":
			if first.Year != second.Year {
				return first.Year > second.Year
			}
		case "horsepower_desc":
			if horsepower(first) != horsepower(second) {
				return horsepower(first) > horsepower(second)
			}
		case "seats_desc":
			if first.Seats != second.Seats {
				return first.Seats > second.Seats
			}
		default:
			firstScore := relevanceScore(first, filters)
			secondScore := relevanceScore(second, filters)
			if firstScore != secondScore {
				return firstScore > secondScore
			}
		}

		return tieBreak(first, second)
	})

	return ranked
}

func BuildAnswer(message string, totalMatches int, returnedCars int, sortBy string) string {
	isUkrainian := IsLikelyUkrainian(message)
	sortBy = normalizeSort(sortBy)

	if totalMatches == 0 {
		if isUkrainian {
			return "На жаль, ми не знайшли автомобілів, які відповідають вашим критеріям."
		}
		return "Unfortunately, we did not find cars that match your criteria."
	}

	if sortBy == "price_asc" {
		if isUkrainian {
			return pluralUkrainianAvailable(totalMatches) + " Нижче показано " + pluralUkrainianCheapest(returnedCars) + "."
		}
		return pluralEnglishAvailable(totalMatches) + " Below are the " + pluralEnglishCheapest(returnedCars) + "."
	}

	if sortBy == "price_desc" {
		if isUkrainian {
			return pluralUkrainianAvailable(totalMatches) + " Нижче показано " + pluralUkrainianMostExpensive(returnedCars) + "."
		}
		return pluralEnglishAvailable(totalMatches) + " Below are the " + pluralEnglishMostExpensive(returnedCars) + "."
	}

	if isUkrainian {
		return "Ми знайшли " + intText(totalMatches) + " " + ukrainianOptionsWord(totalMatches) + ", які відповідають базовим критеріям. Нижче показано " + intText(returnedCars) + " " + ukrainianBestRecommendationsWord(returnedCars) + "."
	}
	return "We found " + intText(totalMatches) + " " + englishOptionsWord(totalMatches) + " that match the basic criteria. Below are the top " + intText(returnedCars) + " " + englishRecommendationsWord(returnedCars) + "."
}

func IsLikelyUkrainian(message string) bool {
	return strings.IndexFunc(message, func(r rune) bool {
		return unicode.In(r, unicode.Cyrillic)
	}) >= 0
}

func relevanceScore(car models.Car, filters models.CarRecommendationFilters) int {
	score := 0

	if containsFold(car.Brand, filters.PreferredBrand) {
		score += 20
	}
	if matchesPreference(car.BodyType, filters.PreferredBodyType) {
		score += 25
	}
	if matchesPreference(car.FuelType, filters.PreferredFuelType) {
		score += 25
	}
	if matchesPreference(car.CarClass, filters.PreferredClass) {
		score += 35
	}
	if matchesPreference(car.Transmission, filters.Transmission) {
		score += 15
	}

	purpose := strings.ToLower(filters.Purpose)
	if containsAny(purpose, []string{"family", "comfort", "travel", "сім", "родин", "комфорт", "подорож"}) {
		if containsAny(strings.ToLower(car.BodyType), []string{"suv", "minivan", "van", "liftback"}) {
			score += 16
		}
		if filters.MinSeats > 0 && car.Seats > filters.MinSeats {
			score += minInt((car.Seats-filters.MinSeats)*3, 12)
		}
		if strings.EqualFold(car.Transmission, "Automatic") {
			score += 8
		}
	}

	if containsAny(purpose, []string{"business", "premium", "бізнес", "преміум"}) {
		if containsAny(strings.ToLower(car.CarClass), []string{"premium", "luxury", "business"}) {
			score += 24
		}
		if containsAny(strings.ToLower(car.BodyType), []string{"sedan", "liftback"}) {
			score += 16
		}
		if filters.MinSeats <= 5 && containsAny(strings.ToLower(car.BodyType), []string{"minivan", "van"}) {
			score -= 14
		}
		if filters.MinSeats <= 5 && strings.Contains(strings.ToLower(car.BodyType), "suv") {
			score -= 6
		}
		if strings.EqualFold(car.Transmission, "Automatic") {
			score += 6
		}
	}

	if containsAny(purpose, []string{"electric", "електро", "електр"}) {
		if strings.EqualFold(car.FuelType, "Electric") {
			score += 30
		}
		if strings.Contains(strings.ToLower(car.CarClass), "electric") {
			score += 10
		}
	}

	return score
}

func tieBreak(first models.Car, second models.Car) bool {
	if first.PricePerDay != second.PricePerDay {
		return first.PricePerDay < second.PricePerDay
	}
	if first.Year != second.Year {
		return first.Year > second.Year
	}
	if horsepower(first) != horsepower(second) {
		return horsepower(first) > horsepower(second)
	}
	return first.ID < second.ID
}

func horsepower(car models.Car) int {
	if car.Horsepower == nil {
		return 0
	}
	return *car.Horsepower
}

func matchesPreference(value string, preference string) bool {
	return preference != "" && containsFold(value, preference)
}

func containsFold(value string, needle string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	needle = strings.ToLower(strings.TrimSpace(needle))
	return needle != "" && strings.Contains(value, needle)
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func normalizeSort(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "price_asc", "price_desc", "year_desc", "horsepower_desc", "seats_desc":
		return strings.ToLower(strings.TrimSpace(sortBy))
	default:
		return "relevance"
	}
}

func minInt(first int, second int) int {
	if first < second {
		return first
	}
	return second
}

func intText(value int) string {
	return strconv.Itoa(value)
}

func pluralEnglishAvailable(count int) string {
	return "We found " + intText(count) + " available " + englishCarsWord(count) + "."
}

func pluralEnglishCheapest(count int) string {
	return intText(count) + " cheapest " + englishOptionsWord(count)
}

func pluralEnglishMostExpensive(count int) string {
	return intText(count) + " most expensive " + englishOptionsWord(count)
}

func pluralEnglishOptionNoun(count int) string {
	if count == 1 {
		return "option"
	}
	return "options"
}

func englishOptionsWord(count int) string {
	return pluralEnglishOptionNoun(count)
}

func englishRecommendationsWord(count int) string {
	if count == 1 {
		return "recommendation"
	}
	return "recommendations"
}

func englishCarsWord(count int) string {
	if count == 1 {
		return "car"
	}
	return "cars"
}

func pluralUkrainianAvailable(count int) string {
	return "Ми знайшли " + intText(count) + " " + ukrainianAvailableCarsWord(count) + "."
}

func pluralUkrainianCheapest(count int) string {
	return intText(count) + " " + ukrainianCheapestOptionsWord(count)
}

func pluralUkrainianMostExpensive(count int) string {
	return intText(count) + " " + ukrainianMostExpensiveOptionsWord(count)
}

func ukrainianAvailableCarsWord(count int) string {
	switch ukrainianPluralForm(count) {
	case 1:
		return "доступний автомобіль"
	case 2:
		return "доступні автомобілі"
	default:
		return "доступних автомобілів"
	}
}

func ukrainianOptionsWord(count int) string {
	switch ukrainianPluralForm(count) {
	case 1:
		return "варіант"
	case 2:
		return "варіанти"
	default:
		return "варіантів"
	}
}

func ukrainianBestRecommendationsWord(count int) string {
	switch ukrainianPluralForm(count) {
	case 1:
		return "найкращу рекомендацію"
	case 2:
		return "найкращі рекомендації"
	default:
		return "найкращих рекомендацій"
	}
}

func ukrainianCheapestOptionsWord(count int) string {
	switch ukrainianPluralForm(count) {
	case 1:
		return "найдешевший варіант"
	case 2:
		return "найдешевші варіанти"
	default:
		return "найдешевших варіантів"
	}
}

func ukrainianMostExpensiveOptionsWord(count int) string {
	switch ukrainianPluralForm(count) {
	case 1:
		return "найдорожчий варіант"
	case 2:
		return "найдорожчі варіанти"
	default:
		return "найдорожчих варіантів"
	}
}

func ukrainianPluralForm(count int) int {
	mod10 := count % 10
	mod100 := count % 100
	if mod10 == 1 && mod100 != 11 {
		return 1
	}
	if mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14) {
		return 2
	}
	return 5
}
