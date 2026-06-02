import { motion } from 'framer-motion';
import { Calendar, Gauge, RefreshCw, Search, Shield, SlidersHorizontal, Users } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { currencyFormatter, displayVehicleTerm, fallbackImageUrl, mainImage, translatableVehicleTerms } from '../lib/carDisplay';
import type { Car } from '../types/api';
import BookingModal from './BookingModal';
import VehicleDetailsModal from './VehicleDetailsModal';

interface ShowroomSectionProps {
  cars: Car[];
  isLoading: boolean;
  error: string;
  token?: string;
  onRetry: () => void;
  onBookingCreated: () => void | Promise<void>;
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

const fieldClass =
  'h-11 w-full rounded-lg border border-cyan-500/25 bg-black/60 px-3 text-sm text-white placeholder-gray-500 focus:border-cyan-300 focus:outline-none transition-colors';

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

const ShowroomSection = ({ cars, isLoading, error, token, onRetry, onBookingCreated }: ShowroomSectionProps) => {
  const [filters, setFilters] = useState<FilterState>(defaultFilters);
  const [isFiltersOpen, setIsFiltersOpen] = useState(true);
  const [selectedCar, setSelectedCar] = useState<Car | null>(null);
  const [bookingCar, setBookingCar] = useState<Car | null>(null);

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

  const { t } = useTranslation([
    'Loading cars...',
    'of',
    'vehicle',
    'vehicles',
    'Filters',
    'Sort showroom',
    'Hide filters',
    'Show filters',
    'Search',
    'Brand, model, color',
    'Class',
    'All classes',
    'Body type',
    'All body types',
    'Transmission',
    'All transmissions',
    'Fuel',
    'All fuel types',
    'Sort by',
    'Price',
    'Year',
    'Horsepower',
    'Engine',
    'Deposit',
    'Seats',
    'Brand',
    'Newest',
    'Order',
    'Ascending',
    'Descending',
    'Reset',
    'Refresh',
    'Active',
    'Unable to load vehicles.',
    'No vehicles match these filters.',
    'Reset sorting or change filters to see more cars.',
    'seats',
    'doors',
    'day',
    'Details',
    'Booking',
    error,
    ...translatableVehicleTerms([
      ...filterOptions.classes,
      ...filterOptions.bodyTypes,
      ...filterOptions.transmissions,
      ...filterOptions.fuelTypes,
      ...visibleCars.flatMap((car) => [car.car_class, car.body_type, car.transmission, car.fuel_type, car.status]),
    ]),
  ]);
  const vehicleCountText = isLoading
    ? t('Loading cars...')
    : `${visibleCars.length} ${t('of')} ${cars.length} ${cars.length === 1 ? t('vehicle') : t('vehicles')}`;
  const hasActiveFilters =
    filters.search.trim() !== '' ||
    filters.carClass !== defaultFilters.carClass ||
    filters.bodyType !== defaultFilters.bodyType ||
    filters.transmission !== defaultFilters.transmission ||
    filters.fuelType !== defaultFilters.fuelType ||
    filters.sortBy !== defaultFilters.sortBy ||
    filters.sortOrder !== defaultFilters.sortOrder;
  const showroomLayoutClass = isFiltersOpen
    ? 'grid min-h-[calc(100vh-4rem)] grid-cols-1 gap-0 lg:grid-cols-[20rem_minmax(0,1fr)]'
    : 'grid min-h-[calc(100vh-4rem)] grid-cols-1 gap-0 lg:grid-cols-[5rem_minmax(0,1fr)]';
  const filterToggleLabel = isFiltersOpen ? t('Hide filters') : t('Show filters');

  const updateFilter = <K extends keyof FilterState>(field: K, value: FilterState[K]) => {
    setFilters((current) => ({
      ...current,
      [field]: value,
    }));
  };

  return (
    <section id="showroom" className="py-0">
      <div className={showroomLayoutClass}>
          <aside
            className={`border-y border-r border-cyan-500/20 bg-black/50 shadow-2xl lg:sticky lg:top-16 lg:min-h-[calc(100vh-4rem)] ${
              isFiltersOpen ? 'p-5' : 'p-3 lg:p-4'
            }`}
          >
            {isFiltersOpen ? (
              <>
            <div className="mb-5 flex items-center justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Filters')}</p>
                <h3 className="mt-1 text-xl font-semibold text-white">{t('Sort showroom')}</h3>
                <p className="mt-2 text-sm text-gray-400">{vehicleCountText}</p>
              </div>
              <button
                type="button"
                onClick={() => setIsFiltersOpen(false)}
                className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-200 transition-colors hover:bg-cyan-500/10"
                aria-controls="showroom-filters-panel"
                aria-expanded={isFiltersOpen}
                aria-label={filterToggleLabel}
                title={filterToggleLabel}
              >
                <SlidersHorizontal size={20} />
              </button>
            </div>

            <div id="showroom-filters-panel" className="space-y-4">
              <label className="relative block">
                <span className="mb-2 block text-sm text-gray-300">{t('Search')}</span>
                <Search size={16} className="absolute left-3 top-[2.65rem] text-cyan-300" />
                <input
                  value={filters.search}
                  onChange={(event) => updateFilter('search', event.target.value)}
                  className={`${fieldClass} pl-10`}
                  placeholder={t('Brand, model, color')}
                />
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Class')}</span>
                <select value={filters.carClass} onChange={(event) => updateFilter('carClass', event.target.value)} className={fieldClass}>
                  <option value="all">{t('All classes')}</option>
                  {filterOptions.classes.map((value) => (
                    <option key={value} value={value}>
                      {displayVehicleTerm(value, t)}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Body type')}</span>
                <select value={filters.bodyType} onChange={(event) => updateFilter('bodyType', event.target.value)} className={fieldClass}>
                  <option value="all">{t('All body types')}</option>
                  {filterOptions.bodyTypes.map((value) => (
                    <option key={value} value={value}>
                      {displayVehicleTerm(value, t)}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Transmission')}</span>
                <select
                  value={filters.transmission}
                  onChange={(event) => updateFilter('transmission', event.target.value)}
                  className={fieldClass}
                >
                  <option value="all">{t('All transmissions')}</option>
                  {filterOptions.transmissions.map((value) => (
                    <option key={value} value={value}>
                      {displayVehicleTerm(value, t)}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Fuel')}</span>
                <select value={filters.fuelType} onChange={(event) => updateFilter('fuelType', event.target.value)} className={fieldClass}>
                  <option value="all">{t('All fuel types')}</option>
                  {filterOptions.fuelTypes.map((value) => (
                    <option key={value} value={value}>
                      {displayVehicleTerm(value, t)}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Sort by')}</span>
                <select
                  value={filters.sortBy}
                  onChange={(event) => updateFilter('sortBy', event.target.value as SortKey)}
                  className={fieldClass}
                >
                  <option value="price_per_day">{t('Price')}</option>
                  <option value="year">{t('Year')}</option>
                  <option value="horsepower">{t('Horsepower')}</option>
                  <option value="engine_volume">{t('Engine')}</option>
                  <option value="deposit">{t('Deposit')}</option>
                  <option value="seats">{t('Seats')}</option>
                  <option value="brand">{t('Brand')}</option>
                  <option value="created_at">{t('Newest')}</option>
                </select>
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Order')}</span>
                <select
                  value={filters.sortOrder}
                  onChange={(event) => updateFilter('sortOrder', event.target.value as SortOrder)}
                  className={fieldClass}
                >
                  <option value="asc">{t('Ascending')}</option>
                  <option value="desc">{t('Descending')}</option>
                </select>
              </label>

              <div className="grid grid-cols-2 gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setFilters(defaultFilters)}
                  className="rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 hover:bg-cyan-500/10 transition-colors"
                >
                  {t('Reset')}
                </button>
                <button
                  type="button"
                  onClick={onRetry}
                  disabled={isLoading}
                  className="inline-flex items-center justify-center gap-2 rounded-lg bg-cyan-500 px-3 py-2 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
                >
                  <RefreshCw size={15} className={isLoading ? 'animate-spin' : ''} />
                  {t('Refresh')}
                </button>
              </div>
            </div>
              </>
            ) : (
              <div className="flex items-center justify-between gap-3 lg:flex-col lg:justify-start">
                <div className="min-w-0 lg:hidden">
                  <p className="text-xs font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Filters')}</p>
                  <p className="mt-2 text-xs text-gray-400 lg:hidden">{vehicleCountText}</p>
                  {hasActiveFilters && (
                    <span className="mt-2 inline-flex rounded-full border border-cyan-300/30 bg-cyan-500/10 px-2 py-1 text-[0.65rem] font-semibold uppercase tracking-wide text-cyan-200">
                      {t('Active')}
                    </span>
                  )}
                </div>
                {hasActiveFilters && <span className="hidden h-2 w-2 rounded-full bg-cyan-300 lg:block" aria-hidden="true" />}
                <button
                  type="button"
                  onClick={() => setIsFiltersOpen(true)}
                  className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-200 transition-colors hover:bg-cyan-500/10"
                  aria-controls="showroom-filters-panel"
                  aria-expanded={isFiltersOpen}
                  aria-label={filterToggleLabel}
                  title={filterToggleLabel}
                >
                  <SlidersHorizontal size={20} />
                </button>
              </div>
            )}
          </aside>

          <div className="min-w-0 px-4 py-5 lg:px-8">
            {error ? (
              <div className="text-center py-10 bg-red-500/10 border border-red-400/20 rounded-xl">
                <p className="text-xl text-red-100">{t('Unable to load vehicles.')}</p>
                <p className="text-sm text-red-200/80 mt-2">{t(error)}</p>
              </div>
            ) : isLoading ? (
              <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" aria-label="Loading vehicles">
                {[0, 1, 2, 3].map((item) => (
                  <div key={item} className="h-96 animate-pulse rounded-xl border border-cyan-500/10 bg-black/40" />
                ))}
              </div>
            ) : visibleCars.length === 0 ? (
              <div className="text-center py-10 bg-black/40 border border-cyan-500/10 rounded-xl">
                <p className="text-xl text-gray-300">{t('No vehicles match these filters.')}</p>
                <p className="text-sm text-gray-400 mt-2">{t('Reset sorting or change filters to see more cars.')}</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
                {visibleCars.map((car) => (
                  <motion.article
                    key={car.id}
                    whileHover={{ scale: 1.02 }}
                    className="overflow-hidden rounded-xl border border-cyan-500/20 bg-black/45 shadow-lg transition-all duration-300 hover:border-cyan-300/50 hover:shadow-cyan-500/25"
                  >
                    <div className="relative h-44">
                      <img
                        src={mainImage(car)}
                        alt={`${car.brand} ${car.model}`}
                        referrerPolicy="no-referrer"
                        className="h-full w-full object-cover"
                        onError={(event) => {
                          event.currentTarget.onerror = null;
                          event.currentTarget.src = fallbackImageUrl;
                        }}
                      />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
                      <span className="absolute top-3 left-3 text-xs px-3 py-1 rounded-full bg-cyan-500/20 border border-cyan-300/30 text-cyan-100">
                        {displayVehicleTerm(car.car_class, t)}
                      </span>
                      <span className="absolute top-3 right-3 text-xs px-3 py-1 rounded-full bg-black/60 border border-white/20 text-gray-100 capitalize">
                        {displayVehicleTerm(car.status, t)}
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
                            {car.year} | {displayVehicleTerm(car.body_type, t)}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Users size={15} className="text-cyan-300" />
                          <span>
                            {car.seats} {t('seats')} | {car.doors} {t('doors')}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Gauge size={15} className="text-cyan-300" />
                          <span>
                            {displayVehicleTerm(car.transmission, t)} | {displayVehicleTerm(car.fuel_type, t)}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Shield size={15} className="text-cyan-300" />
                          <span>
                            {t('Deposit')} {currencyFormatter.format(car.deposit)}
                          </span>
                        </div>
                      </div>
                      <p className="font-semibold text-cyan-300">
                        {currencyFormatter.format(car.price_per_day)} / {t('day')}
                      </p>
                      <div className="mt-4 grid grid-cols-2 gap-3">
                        <button
                          type="button"
                          onClick={() => setSelectedCar(car)}
                          className="rounded-lg border border-cyan-500/30 px-4 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                        >
                          {t('Details')}
                        </button>
                        <button
                          type="button"
                          onClick={() => setBookingCar(car)}
                          className="rounded-lg bg-cyan-500 px-4 py-2 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
                        >
                          {t('Booking')}
                        </button>
                      </div>
                    </div>
                  </motion.article>
                ))}
              </div>
            )}
          </div>
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
      {bookingCar && <BookingModal car={bookingCar} token={token} onCreated={onBookingCreated} onClose={() => setBookingCar(null)} />}
    </section>
  );
};

export default ShowroomSection;
