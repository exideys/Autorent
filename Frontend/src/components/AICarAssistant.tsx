import { Calendar, DollarSign, Fuel, Gauge, Loader2, Send, Shield, Sparkles, Users } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { getCarRecommendation } from '../lib/api';
import { displayVehicleTerm, translatableVehicleTerms } from '../lib/carDisplay';
import type { CarRecommendationResponse, RecommendedCar } from '../types/api';

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

const requestPlaceholder =
  'I need a comfortable SUV for 5 people under 240 dollars per day\npremium business car for 5 people up to 240 dollars per day';

const fieldClass =
  'min-h-28 w-full resize-y rounded-lg border border-cyan-500/25 bg-black/60 px-4 py-3 text-sm text-white placeholder-gray-500 focus:border-cyan-300 focus:outline-none transition-colors';

const recommendationDetail = (car: RecommendedCar, t: (text: string) => string) => [
  { icon: Calendar, label: `${car.year}` },
  { icon: Users, label: `${car.seats} ${t('seats')}` },
  { icon: Gauge, label: car.horsepower > 0 ? `${car.horsepower} ${t('hp')}` : t('Horsepower n/a') },
  { icon: Shield, label: `${t('Deposit')} ${currencyFormatter.format(car.deposit)}` },
  { icon: Fuel, label: displayVehicleTerm(car.fuel_type, t) },
];

const AICarAssistant = () => {
  const [message, setMessage] = useState('');
  const [result, setResult] = useState<CarRecommendationResponse | null>(null);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { t } = useTranslation([
    requestPlaceholder,
    'Message is required.',
    'Unable to get recommendations.',
    'AI Car Assistant',
    'Find the right rental faster',
    'match',
    'matches',
    'Car request',
    'Searching...',
    'Find cars',
    'available',
    'day',
    'seats',
    'hp',
    'Horsepower n/a',
    'Deposit',
    error,
    result?.answer,
    ...translatableVehicleTerms(result?.cars.flatMap((car) => [car.car_class, car.status || 'available', car.body_type, car.transmission, car.fuel_type]) || []),
  ]);

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
              {t('AI Car Assistant')}
            </p>
            <h2 className="mt-2 text-2xl font-semibold text-white md:text-3xl">{t('Find the right rental faster')}</h2>
          </div>
          {result && (
            <p className="rounded-lg border border-emerald-400/25 bg-emerald-500/10 px-3 py-2 text-sm font-semibold text-emerald-100">
              {result.total_matches} {result.total_matches === 1 ? t('match') : t('matches')}
            </p>
          )}
        </div>

        <form onSubmit={handleSubmit} className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-start">
          <label className="block">
            <span className="sr-only">{t('Car request')}</span>
            <textarea
              value={message}
              onChange={(event) => setMessage(event.target.value)}
              className={fieldClass}
              placeholder={t(requestPlaceholder)}
              disabled={isLoading}
            />
          </label>
          <button
            type="submit"
            disabled={isLoading}
            className="inline-flex min-h-12 items-center justify-center gap-2 rounded-lg bg-cyan-500 px-5 py-3 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors lg:min-w-40"
          >
            {isLoading ? <Loader2 size={17} className="animate-spin" /> : <Send size={17} />}
            {isLoading ? t('Searching...') : t('Find cars')}
          </button>
        </form>

        {error && (
          <div className="mt-4 rounded-lg border border-red-400/30 bg-red-500/10 px-4 py-3 text-sm text-red-100" role="alert">
            {t(error)}
          </div>
        )}

        {result && (
          <div className="mt-6 space-y-5" aria-live="polite">
            <div className="rounded-lg border border-cyan-500/20 bg-cyan-500/10 px-4 py-3 text-sm text-cyan-50">
              {t(result.answer)}
            </div>

            {result.cars.length > 0 && (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5">
                {result.cars.map((car) => (
                  <article key={car.id} className="rounded-lg border border-cyan-500/20 bg-black/45 p-4 shadow-lg">
                    <div className="mb-3 flex items-start justify-between gap-3">
                      <div>
                        <h3 className="text-lg font-semibold leading-tight text-white">
                          {car.brand} {car.model}
                        </h3>
                        <p className="mt-1 text-xs uppercase tracking-wide text-cyan-200">{displayVehicleTerm(car.car_class, t)}</p>
                      </div>
                      <span className="rounded-md border border-white/15 bg-white/10 px-2 py-1 text-xs capitalize text-gray-200">
                        {displayVehicleTerm(car.status || 'available', t)}
                      </span>
                    </div>

                    <div className="mb-4 flex flex-wrap gap-2 text-xs text-gray-300">
                      <span className="rounded-md bg-white/10 px-2 py-1">{displayVehicleTerm(car.body_type, t)}</span>
                      <span className="rounded-md bg-white/10 px-2 py-1">{displayVehicleTerm(car.transmission, t)}</span>
                      <span className="rounded-md bg-white/10 px-2 py-1">{displayVehicleTerm(car.fuel_type, t)}</span>
                    </div>

                    <p className="mb-4 inline-flex items-center gap-2 text-xl font-bold text-cyan-300">
                      <DollarSign size={18} />
                      {currencyFormatter.format(car.price_per_day)} / {t('day')}
                    </p>

                    <div className="space-y-2 text-sm text-gray-300">
                      {recommendationDetail(car, t).map(({ icon: Icon, label }) => (
                        <div key={label} className="flex items-center gap-2">
                          <Icon size={15} className="text-cyan-300" />
                          <span>{label}</span>
                        </div>
                      ))}
                    </div>
                  </article>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </section>
  );
};

export default AICarAssistant;
