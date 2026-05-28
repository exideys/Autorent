import { motion } from 'framer-motion';
import { Calendar, Gauge, RefreshCw, Search, Shield, SlidersHorizontal, Users, X } from 'lucide-react';
import { useMemo, useState, type KeyboardEvent } from 'react';
import type { Car } from '../types/api';

interface ShowroomSectionProps {
  cars: Car[];
  isLoading: boolean;
  error: string;
  onRetry: () => void;
}

type SortKey =
  | 'price_per_day'
  | 'year'
  | 'horsepower'
  | 'engine_volume'
  | 'deposit'
  | 'seats'
  | 'brand'
  | 'created_at';

type SortOrder = 'asc' | 'desc';

interface FilterState {
  search: string;
  carClass: string;
  bodyType: string;
  transmission: string;
  fuelType: string;
  sortBy: SortKey;
  sortOrder: SortOrder;
}

const defaultFilters: FilterState = {
  search: '',
  carClass: 'all',
  bodyType: 'all',
  transmission: 'all',
  fuelType: 'all',
  sortBy: 'price_per_day',
  sortOrder: 'asc',
};

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

const fieldClass =
  'h-11 w-full rounded-lg border border-cyan-500/25 bg-black/60 px-3 text-sm text-white placeholder-gray-500 focus:border-cyan-300 focus:outline-none transition-colors';

const fallbackImageUrl = `${import.meta.env.BASE_URL}hero-main.png`;

const mainImage = (car: Car) => {
  const selectedImage = car.images.find((image) => image.is_main) || car.images[0];
  return selectedImage?.image_url || fallbackImageUrl;
};

const uniqueValues = (cars: Car[], selector: (car: Car) => string) =>
  Array.from(new Set(cars.map(selector).filter(Boolean))).sort((first, second) => first.localeCompare(second));

const numericValue = (car: Car, sortBy: SortKey) => {
  switch (sortBy) {
    case 'price_per_day':
      return car.price_per_day;
    case 'year':
      return car.year;
    case 'horsepower':
      return car.horsepower || 0;
    case 'engine_volume':
      return car.engine_volume || 0;
    case 'deposit':
      return car.deposit;
    case 'seats':
      return car.seats;
    case 'created_at':
      return new Date(car.created_at).getTime() || 0;
    default:
      return 0;
  }
};

const matchesText = (car: Car, search: string) => {
  const normalizedSearch = search.trim().toLowerCase();
  if (!normalizedSearch) {
    return true;
  }

  return [car.brand, car.model, car.car_class, car.body_type, car.transmission, car.fuel_type, car.color || '', car.year.toString()]
    .join(' ')
    .toLowerCase()
    .includes(normalizedSearch);
};

const detailRows = (car: Car) => [
  ['Brand', car.brand],
  ['Model', car.model],
  ['Year', car.year],
  ['Class', car.car_class],
  ['Body type', car.body_type],
  ['Transmission', car.transmission],
  ['Fuel type', car.fuel_type],
  ['Seats', car.seats],
  ['Doors', car.doors],
  ['Engine volume', car.engine_volume ? `${car.engine_volume}L` : 'Not specified'],
  ['Horsepower', car.horsepower ? `${car.horsepower} hp` : 'Not specified'],
  ['Price per day', currencyFormatter.format(car.price_per_day)],
  ['Deposit', currencyFormatter.format(car.deposit)],
  ['Color', car.color || 'Not specified'],
  ['Status', car.status],
];

const ShowroomSection = ({ cars, isLoading, error, onRetry }: ShowroomSectionProps) => {
  const [filters, setFilters] = useState<FilterState>(defaultFilters);
  const [selectedCar, setSelectedCar] = useState<Car | null>(null);

  const filterOptions = useMemo(
    () => ({
      classes: uniqueValues(cars, (car) => car.car_class),
      bodyTypes: uniqueValues(cars, (car) => car.body_type),
      transmissions: uniqueValues(cars, (car) => car.transmission),
      fuelTypes: uniqueValues(cars, (car) => car.fuel_type),
    }),
    [cars],
  );

  const visibleCars = useMemo(() => {
    const filteredCars = cars.filter((car) => {
      if (!matchesText(car, filters.search)) {
        return false;
      }
      if (filters.carClass !== 'all' && car.car_class !== filters.carClass) {
        return false;
      }
      if (filters.bodyType !== 'all' && car.body_type !== filters.bodyType) {
        return false;
      }
      if (filters.transmission !== 'all' && car.transmission !== filters.transmission) {
        return false;
      }
      if (filters.fuelType !== 'all' && car.fuel_type !== filters.fuelType) {
        return false;
      }
      return true;
    });

    return [...filteredCars].sort((firstCar, secondCar) => {
      if (filters.sortBy === 'brand') {
        const comparison = `${firstCar.brand} ${firstCar.model}`.localeCompare(`${secondCar.brand} ${secondCar.model}`);
        return filters.sortOrder === 'asc' ? comparison : -comparison;
      }

      const comparison = numericValue(firstCar, filters.sortBy) - numericValue(secondCar, filters.sortBy);
      return filters.sortOrder === 'asc' ? comparison : -comparison;
    });
  }, [cars, filters]);

  const updateFilter = <K extends keyof FilterState>(field: K, value: FilterState[K]) => {
    setFilters((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const openWithKeyboard = (event: KeyboardEvent<HTMLElement>, car: Car) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      setSelectedCar(car);
    }
  };

  return (
    <section id="showroom" className="py-0">
      <div className="grid min-h-[calc(100vh-4rem)] grid-cols-1 gap-0 lg:grid-cols-[20rem_minmax(0,1fr)]">
          <aside className="border-y border-r border-cyan-500/20 bg-black/50 p-5 shadow-2xl lg:sticky lg:top-16 lg:min-h-[calc(100vh-4rem)]">
            <div className="mb-5 flex items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-cyan-300">Filters</p>
                <h3 className="mt-1 text-xl font-semibold text-white">Sort showroom</h3>
                <p className="mt-2 text-sm text-gray-400">
                  {isLoading
                    ? 'Loading cars...'
                    : `${visibleCars.length} of ${cars.length} vehicle${cars.length === 1 ? '' : 's'}`}
                </p>
              </div>
              <SlidersHorizontal size={20} className="text-cyan-300" />
            </div>

            <div className="space-y-4">
              <label className="relative block">
                <span className="mb-2 block text-sm text-gray-300">Search</span>
                <Search size={16} className="absolute left-3 top-[2.65rem] text-cyan-300" />
                <input
                  value={filters.search}
                  onChange={(event) => updateFilter('search', event.target.value)}
                  className={`${fieldClass} pl-10`}
                  placeholder="Brand, model, color"
                />
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>Class</span>
                <select value={filters.carClass} onChange={(event) => updateFilter('carClass', event.target.value)} className={fieldClass}>
                  <option value="all">All classes</option>
                  {filterOptions.classes.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>Body type</span>
                <select value={filters.bodyType} onChange={(event) => updateFilter('bodyType', event.target.value)} className={fieldClass}>
                  <option value="all">All body types</option>
                  {filterOptions.bodyTypes.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>Transmission</span>
                <select
                  value={filters.transmission}
                  onChange={(event) => updateFilter('transmission', event.target.value)}
                  className={fieldClass}
                >
                  <option value="all">All transmissions</option>
                  {filterOptions.transmissions.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>Fuel</span>
                <select value={filters.fuelType} onChange={(event) => updateFilter('fuelType', event.target.value)} className={fieldClass}>
                  <option value="all">All fuel types</option>
                  {filterOptions.fuelTypes.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>Sort by</span>
                <select
                  value={filters.sortBy}
                  onChange={(event) => updateFilter('sortBy', event.target.value as SortKey)}
                  className={fieldClass}
                >
                  <option value="price_per_day">Price</option>
                  <option value="year">Year</option>
                  <option value="horsepower">Horsepower</option>
                  <option value="engine_volume">Engine</option>
                  <option value="deposit">Deposit</option>
                  <option value="seats">Seats</option>
                  <option value="brand">Brand</option>
                  <option value="created_at">Newest</option>
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>Order</span>
                <select
                  value={filters.sortOrder}
                  onChange={(event) => updateFilter('sortOrder', event.target.value as SortOrder)}
                  className={fieldClass}
                >
                  <option value="asc">Ascending</option>
                  <option value="desc">Descending</option>
                </select>
              </label>

              <div className="grid grid-cols-2 gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setFilters(defaultFilters)}
                  className="rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 hover:bg-cyan-500/10 transition-colors"
                >
                  Reset
                </button>
                <button
                  type="button"
                  onClick={onRetry}
                  disabled={isLoading}
                  className="inline-flex items-center justify-center gap-2 rounded-lg bg-cyan-500 px-3 py-2 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
                >
                  <RefreshCw size={15} className={isLoading ? 'animate-spin' : ''} />
                  Refresh
                </button>
              </div>
            </div>
          </aside>

          <div className="min-w-0 px-4 py-5 lg:px-8">
            {error ? (
              <div className="text-center py-10 bg-red-500/10 border border-red-400/20 rounded-xl">
                <p className="text-xl text-red-100">Unable to load vehicles.</p>
                <p className="text-sm text-red-200/80 mt-2">{error}</p>
              </div>
            ) : isLoading ? (
              <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" aria-label="Loading vehicles">
                {[0, 1, 2, 3].map((item) => (
                  <div key={item} className="h-96 animate-pulse rounded-xl border border-cyan-500/10 bg-black/40" />
                ))}
              </div>
            ) : visibleCars.length === 0 ? (
              <div className="text-center py-10 bg-black/40 border border-cyan-500/10 rounded-xl">
                <p className="text-xl text-gray-300">No vehicles match these filters.</p>
                <p className="text-sm text-gray-400 mt-2">Reset sorting or change filters to see more cars.</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                {visibleCars.map((car) => (
                  <motion.article
                    key={car.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => setSelectedCar(car)}
                    onKeyDown={(event) => openWithKeyboard(event, car)}
                    whileHover={{ scale: 1.02 }}
                    className="cursor-pointer overflow-hidden rounded-xl border border-cyan-500/20 bg-black/45 shadow-lg hover:border-cyan-300/50 hover:shadow-cyan-500/25 focus:border-cyan-300 focus:outline-none transition-all duration-300"
                  >
                    <div className="relative h-44">
                      <img
                        src={mainImage(car)}
                        alt={`${car.brand} ${car.model}`}
                        className="h-full w-full object-cover"
                        onError={(event) => {
                          event.currentTarget.onerror = null;
                          event.currentTarget.src = fallbackImageUrl;
                        }}
                      />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
                      <span className="absolute top-3 left-3 text-xs px-3 py-1 rounded-full bg-cyan-500/20 border border-cyan-300/30 text-cyan-100">
                        {car.car_class}
                      </span>
                      <span className="absolute top-3 right-3 text-xs px-3 py-1 rounded-full bg-black/60 border border-white/20 text-gray-100 capitalize">
                        {car.status}
                      </span>
                    </div>
                    <div className="p-5">
                      <h3 className="text-xl font-semibold mb-2">
                        {car.brand} {car.model}
                      </h3>
                      <div className="space-y-2 text-sm text-gray-300 mb-4">
                        <div className="flex items-center gap-2">
                          <Calendar size={15} className="text-cyan-300" />
                          <span>
                            {car.year} | {car.body_type}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Users size={15} className="text-cyan-300" />
                          <span>
                            {car.seats} seats | {car.doors} doors
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Gauge size={15} className="text-cyan-300" />
                          <span>
                            {car.transmission} | {car.fuel_type}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Shield size={15} className="text-cyan-300" />
                          <span>Deposit {currencyFormatter.format(car.deposit)}</span>
                        </div>
                      </div>
                      <div className="flex items-center justify-between gap-3">
                        <p className="text-cyan-300 font-semibold">{currencyFormatter.format(car.price_per_day)} / day</p>
                        <span className="rounded-lg bg-cyan-500 px-4 py-2 text-sm font-semibold text-black">Details</span>
                      </div>
                    </div>
                  </motion.article>
                ))}
              </div>
            )}
          </div>
      </div>

      {selectedCar && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/75 px-4 py-6 backdrop-blur-sm">
          <motion.div
            initial={{ opacity: 0, y: 24, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            className="max-h-[90vh] w-full max-w-5xl overflow-y-auto rounded-xl border border-cyan-500/30 bg-gray-950 shadow-2xl"
          >
            <div className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-cyan-500/20 bg-gray-950/95 p-5 backdrop-blur">
              <div>
                <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">Vehicle details</p>
                <h3 className="mt-1 text-2xl font-bold text-white">
                  {selectedCar.brand} {selectedCar.model}
                </h3>
              </div>
              <button
                type="button"
                onClick={() => setSelectedCar(null)}
                className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-100 hover:bg-cyan-500/10 transition-colors"
                aria-label="Close vehicle details"
              >
                <X size={20} />
              </button>
            </div>

            <div className="grid gap-6 p-5 lg:grid-cols-[minmax(0,1.05fr)_minmax(20rem,0.95fr)]">
              <div className="space-y-4">
                <img
                  src={mainImage(selectedCar)}
                  alt={`${selectedCar.brand} ${selectedCar.model}`}
                  className="h-72 w-full rounded-xl object-cover"
                  onError={(event) => {
                    event.currentTarget.onerror = null;
                    event.currentTarget.src = fallbackImageUrl;
                  }}
                />
                {selectedCar.images.length > 1 && (
                  <div className="grid grid-cols-3 gap-3">
                    {selectedCar.images.slice(0, 6).map((image) => (
                      <img
                        key={image.id}
                        src={image.image_url}
                        alt={`${selectedCar.brand} ${selectedCar.model}`}
                        className="h-24 rounded-lg object-cover"
                        onError={(event) => {
                          event.currentTarget.style.display = 'none';
                        }}
                      />
                    ))}
                  </div>
                )}
              </div>

              <div className="space-y-5">
                <div className="grid grid-cols-2 gap-3">
                  <div className="rounded-xl border border-cyan-500/20 bg-black/35 p-4">
                    <p className="text-sm text-gray-400">Daily price</p>
                    <p className="mt-1 text-2xl font-bold text-cyan-300">{currencyFormatter.format(selectedCar.price_per_day)}</p>
                  </div>
                  <div className="rounded-xl border border-cyan-500/20 bg-black/35 p-4">
                    <p className="text-sm text-gray-400">Deposit</p>
                    <p className="mt-1 text-2xl font-bold text-white">{currencyFormatter.format(selectedCar.deposit)}</p>
                  </div>
                </div>

                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {detailRows(selectedCar).map(([label, value]) => (
                    <div key={label} className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                      <p className="text-xs uppercase tracking-wide text-gray-500">{label}</p>
                      <p className="mt-1 text-sm font-semibold text-gray-100 capitalize">{value}</p>
                    </div>
                  ))}
                </div>

                <button
                  type="button"
                  className="w-full rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-5 py-3 font-semibold text-white shadow-lg shadow-cyan-500/20 hover:from-cyan-600 hover:to-violet-700 transition-colors"
                >
                  Start Booking
                </button>
              </div>
            </div>
          </motion.div>
        </div>
      )}
    </section>
  );
};

export default ShowroomSection;
