export type Car = {
  id: number;
  brand: string;
  model: string;
  year: number;
  car_class: string;
  body_type: string;
  transmission: string;
  fuel_type: string;
  seats: number;
  doors: number;
  engine_volume?: number | null;
  horsepower?: number;
  price_per_day: number;
  deposit: number;
  color?: string;
  status: string;
};

export type CarRecommendationResponse = {
  answer: string;
  cars: Car[];
  total_matches: number;
};

type ApiErrorResponse = {
  error?: string;
};

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || "/api").replace(/\/$/, "");

export async function recommendCar(message: string): Promise<CarRecommendationResponse> {
  const trimmedMessage = message.trim();

  if (!trimmedMessage) {
    throw new Error("Please enter your car rental request.");
  }

  const response = await fetch(`${API_BASE_URL}/ai/car-recommendation`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      message: trimmedMessage,
    }),
  });

  if (!response.ok) {
    let errorMessage = "Failed to get car recommendation.";

    try {
      const errorData = (await response.json()) as ApiErrorResponse;

      if (errorData.error) {
        errorMessage = errorData.error;
      }
    } catch {
      // Keep default error message if response is not JSON.
    }

    throw new Error(errorMessage);
  }

  return response.json() as Promise<CarRecommendationResponse>;
}
