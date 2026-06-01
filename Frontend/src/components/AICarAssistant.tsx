import { Calendar, DollarSign, Fuel, Gauge, Info, Loader2, Send, Shield, Sparkles, Users } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { getCarRecommendation } from '../lib/api';
import { fallbackImageUrl, mainImage } from '../lib/carDisplay';
import type { Car, CarRecommendationResponse, RecommendedCar } from '../types/api';
import BookingModal from './BookingModal';
import VehicleDetailsModal from './VehicleDetailsModal';

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

const requestPlaceholder =
  'I need a comfortable SUV for 5 people under 240 dollars per day\npremium business car for 5 people up to 240 dollars per day';

const fieldClass =
  'min-h-28 w-full resize-y rounded-lg border border-cyan-500/25 bg-black/60 px-4 py-3 text-sm text-white placeholder-gray-500 focus:border-cyan-300 focus:outline-none transition-colors';

const recommendationDetail = (car: RecommendedCar) => [
  { icon: Calendar, label: `${car.year}` },
  { icon: Users, label: `${car.seats} seats` },
  { icon: Gauge, label: car.horsepower > 0 ? `${car.horsepower} hp` : 'Horsepower n/a' },
  { icon: Shield, label: `Deposit ${currencyFormatter.format(car.deposit)}` },
  { icon: Fuel, label: car.fuel_type },
];

const recommendedToCar = (car: RecommendedCar): Car => ({
  id: car.id,
  brand: car.brand,
  model: car.model,
  year: car.year,
  car_class: car.car_class,
  body_type: car.body_type,
  transmission: car.transmission,
  fuel_type: car.fuel_type,
  seats: car.seats,
  doors: car.doors,
  engine_volume: car.engine_volume,
  horsepower: car.horsepower || undefined,
  price_per_day: car.price_per_day,
  deposit: car.deposit,
  color: car.color || undefined,
  status: car.status,
  created_at: car.created_at || new Date().toISOString(),
  images: car.images || [],
});

const AICarAssistant = () => {
  const [message, setMessage] = useState('');
  const [result, setResult] = useState<CarRecommendationResponse | null>(null);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [selectedCar, setSelectedCar] = useState<Car | null>(null);
  const [bookingCar, setBookingCar] = useState<Car | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedMessage = message.trim();

    if (!trimmedMessage) {
      setError('Message is required.');
      setResult(null);
      return;
    }

    setIsLoading(true);
    setError('');

    try {
      const response = await getCarRecommendation(trimmedMessage);
      setResult(response);
    } catch (requestError) {
      setResult(null);
      setError(requestError instanceof Error ? requestError.message : 'Unable to get recommendations.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <section className="border-y border-cyan-500/15 bg-black/30 px-4 py-10">
      <div className="mx-auto max-w-7xl">
        <div className="mb-5 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-cyan-300">
              <Sparkles size={16} />
              AI Car Assistant
            </p>
            <h2 className="mt-2 text-2xl font-semibold text-white md:text-3xl">Find the right rental faster</h2>
          </div>
          {result && (
            <p className="rounded-lg border border-emerald-400/25 bg-emerald-500/10 px-3 py-2 text-sm font-semibold text-emerald-100">
              {result.total_matches} match{result.total_matches === 1 ? '' : 'es'}
            </p>
          )}
        </div>

        <form onSubmit={handleSubmit} className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-start">
          <label className="block">
            <span className="sr-only">Car request</span>
            <textarea
              value={message}
              onChange={(event) => setMessage(event.target.value)}
              className={fieldClass}
              placeholder={requestPlaceholder}
              disabled={isLoading}
            />
          </label>
          <button
            type="submit"
            disabled={isLoading}
            className="inline-flex min-h-12 items-center justify-center gap-2 rounded-lg bg-cyan-500 px-5 py-3 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors lg:min-w-40"
          >
            {isLoading ? <Loader2 size={17} className="animate-spin" /> : <Send size={17} />}
            {isLoading ? 'Searching...' : 'Find cars'}
          </button>
        </form>

        {error && (
          <div className="mt-4 rounded-lg border border-red-400/30 bg-red-500/10 px-4 py-3 text-sm text-red-100" role="alert">
            {error}
          </div>
        )}

        {result && (
          <div className="mt-6 space-y-5" aria-live="polite">
            <div className="rounded-lg border border-cyan-500/20 bg-cyan-500/10 px-4 py-3 text-sm text-cyan-50">
              {result.answer}
            </div>

            {result.cars.length > 0 && (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5">
                {result.cars.map((car) => {
                  const displayCar = recommendedToCar(car);

                  return (
                    <article key={car.id} className="overflow-hidden rounded-lg border border-cyan-500/20 bg-black/45 shadow-lg">
                      <div className="relative h-36">
                        <img
                          src={mainImage(displayCar)}
                          alt={`${car.brand} ${car.model}`}
                          referrerPolicy="no-referrer"
                          className="h-full w-full object-cover"
                          onError={(event) => {
                            event.currentTarget.onerror = null;
                            event.currentTarget.src = fallbackImageUrl;
                          }}
                        />
                        <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
                        <span className="absolute left-3 top-3 rounded-md border border-cyan-300/30 bg-cyan-500/20 px-2 py-1 text-xs text-cyan-100">
                          {car.car_class}
                        </span>
                      </div>

                      <div className="p-4">
                        <div className="mb-3 flex items-start justify-between gap-3">
                          <div>
                            <h3 className="text-lg font-semibold leading-tight text-white">
                              {car.brand} {car.model}
                            </h3>
                            <p className="mt-1 text-xs uppercase tracking-wide text-cyan-200">{car.body_type}</p>
                          </div>
                          <span className="rounded-md border border-white/15 bg-white/10 px-2 py-1 text-xs capitalize text-gray-200">
                            {car.status || 'available'}
                          </span>
                        </div>

                        <div className="mb-4 flex flex-wrap gap-2 text-xs text-gray-300">
                          <span className="rounded-md bg-white/10 px-2 py-1">{car.transmission}</span>
                          <span className="rounded-md bg-white/10 px-2 py-1">{car.fuel_type}</span>
                        </div>

                        <p className="mb-4 inline-flex items-center gap-2 text-xl font-bold text-cyan-300">
                          <DollarSign size={18} />
                          {currencyFormatter.format(car.price_per_day)} / day
                        </p>

                        <div className="space-y-2 text-sm text-gray-300">
                          {recommendationDetail(car).map(({ icon: Icon, label }) => (
                            <div key={label} className="flex items-center gap-2">
                              <Icon size={15} className="text-cyan-300" />
                              <span>{label}</span>
                            </div>
                          ))}
                        </div>

                        <button
                          type="button"
                          onClick={() => setSelectedCar(displayCar)}
                          className="mt-4 inline-flex w-full items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-4 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                        >
                          <Info size={16} />
                          Details
                        </button>
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
          </div>
        )}
      </div>

      {selectedCar && (
        <VehicleDetailsModal
          car={selectedCar}
          onClose={() => setSelectedCar(null)}
          onBooking={(car) => {
            setSelectedCar(null);
            setBookingCar(car);
          }}
        />
      )}
      {bookingCar && <BookingModal car={bookingCar} onClose={() => setBookingCar(null)} />}
    </section>
  );
};

export default AICarAssistant;
