import { Car as CarIcon, Edit3, Newspaper, Plus, RefreshCw, Save, Star, Trash2, Users, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState, type ChangeEvent, type FormEvent } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { displayVehicleTerm, translatableVehicleTerms } from '../lib/carDisplay';
import { ApiError, createAdminCar, deleteAdminCar, listAdminCars, listAdminUsers, rateAdminUser, updateAdminCar } from '../lib/api';
import type { Car, CarInput, User } from '../types/api';
import AdminNewsDashboard from './AdminNewsDashboard';

interface AdminDashboardProps {
  token: string;
  onInventoryChanged: () => void;
  onNewsChanged: () => void;
  onUnauthorized: () => void;
}

type AdminPanel = 'fleet' | 'news';

interface CarFormState {
  brand: string;
  model: string;
  year: string;
  carClass: string;
  bodyType: string;
  transmission: string;
  fuelType: string;
  seats: string;
  doors: string;
  engineVolume: string;
  horsepower: string;
  pricePerDay: string;
  deposit: string;
  color: string;
  status: string;
  imageUrls: string;
}

const emptyForm: CarFormState = {
  brand: '',
  model: '',
  year: new Date().getFullYear().toString(),
  carClass: '',
  bodyType: '',
  transmission: 'Automatic',
  fuelType: 'Petrol',
  seats: '5',
  doors: '4',
  engineVolume: '',
  horsepower: '',
  pricePerDay: '',
  deposit: '',
  color: '',
  status: 'available',
  imageUrls: '',
};

const statusOptions = ['available', 'rented', 'maintenance', 'unavailable'];
const ratingOptions = ['1', '1.5', '2', '2.5', '3', '3.5', '4', '4.5', '5'];

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
});

const inputClass =
  'w-full bg-black/60 border border-cyan-500/25 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:border-cyan-400 focus:outline-none transition-colors';

const labelClass = 'block space-y-2 text-sm text-gray-300';

const fallbackImageUrl = `${import.meta.env.BASE_URL}hero-main.png`;

const mainImage = (car: Car) => {
  const selectedImage = car.images.find((image) => image.is_main) || car.images[0];
  return selectedImage?.image_url || fallbackImageUrl;
};

const optionalNumber = (value: string) => {
  const trimmed = value.trim();
  return trimmed === '' ? undefined : Number(trimmed);
};

const formatDate = (value: string) => {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Not available' : dateFormatter.format(date);
};

const imageInputs = (value: string) =>
  value
    .split(/\r?\n/)
    .map((url) => url.trim())
    .filter(Boolean)
    .map((imageUrl, index) => ({
      image_url: imageUrl,
      is_main: index === 0,
      sort_order: index,
    }));

const toCarInput = (form: CarFormState): CarInput => ({
  brand: form.brand.trim(),
  model: form.model.trim(),
  year: Number(form.year),
  car_class: form.carClass.trim(),
  body_type: form.bodyType.trim(),
  transmission: form.transmission.trim(),
  fuel_type: form.fuelType.trim(),
  seats: Number(form.seats),
  doors: Number(form.doors),
  engine_volume: optionalNumber(form.engineVolume),
  horsepower: optionalNumber(form.horsepower),
  price_per_day: Number(form.pricePerDay),
  deposit: Number(form.deposit),
  color: form.color.trim() || undefined,
  status: form.status.trim() || 'available',
  images: imageInputs(form.imageUrls),
});

const formFromCar = (car: Car): CarFormState => ({
  brand: car.brand,
  model: car.model,
  year: car.year.toString(),
  carClass: car.car_class,
  bodyType: car.body_type,
  transmission: car.transmission,
  fuelType: car.fuel_type,
  seats: car.seats.toString(),
  doors: car.doors.toString(),
  engineVolume: car.engine_volume?.toString() || '',
  horsepower: car.horsepower?.toString() || '',
  pricePerDay: car.price_per_day.toString(),
  deposit: car.deposit.toString(),
  color: car.color || '',
  status: car.status || 'available',
  imageUrls: car.images.map((image) => image.image_url).join('\n'),
});

const AdminDashboard = ({ token, onInventoryChanged, onNewsChanged, onUnauthorized }: AdminDashboardProps) => {
  const [cars, setCars] = useState<Car[]>([]);
  const [customers, setCustomers] = useState<User[]>([]);
  const [form, setForm] = useState<CarFormState>(emptyForm);
  const [ratingInputs, setRatingInputs] = useState<Record<number, string>>({});
  const [editingId, setEditingId] = useState<number | null>(null);
  const [activePanel, setActivePanel] = useState<AdminPanel>('fleet');
  const [isLoading, setIsLoading] = useState(true);
  const [isCustomersLoading, setIsCustomersLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const { t } = useTranslation([
    'Not available',
    'Admin',
    'Fleet Dashboard',
    'Manage the live vehicle inventory served from the existing AutoRent backend.',
    'Refresh',
    'Total Vehicles',
    'Available',
    'Average Daily Price',
    'Status Breakdown',
    'No vehicles yet.',
    'Customers',
    'Client ratings',
    'Leave a rating for registered customers. New ratings update their average score.',
    'client',
    'clients',
    'No registered clients yet.',
    'Rating',
    'Count',
    'Joined',
    'Rate',
    'Customer rating must be between 1 and 5.',
    'Rating added.',
    'Unable to rate customer',
    'Vehicle updated.',
    'Vehicle created.',
    'Vehicle deleted.',
    'Unable to save vehicle',
    'Unable to delete vehicle',
    'Delete',
    'Edit Vehicle',
    'Add Vehicle',
    'All required fields map directly to the backend car payload.',
    'Brand',
    'Model',
    'Year',
    'Car Class',
    'Body Type',
    'Transmission',
    'Fuel Type',
    'Status',
    'Seats',
    'Doors',
    'Engine Volume',
    'Horsepower',
    'Price Per Day',
    'Deposit',
    'Color',
    'Image URLs',
    'Optional',
    'One URL per line. The first URL becomes the main image.',
    'Saving...',
    'Save Changes',
    'Create Vehicle',
    'Inventory',
    'Create, edit, and delete cars from the admin API.',
    'No cars in inventory.',
    'Use the form to add the first vehicle.',
    'seats',
    'doors',
    'day',
    'Edit',
    'Cancel editing',
    'Fleet',
    'News',
    error,
    message,
    ...statusOptions,
    ...translatableVehicleTerms(cars.flatMap((car) => [car.status, car.car_class, car.body_type, car.transmission, car.fuel_type])),
    ...customers.map((customer) => customer.status || 'unknown'),
  ]);
  const displayDate = (value: string) => {
    const formattedDate = formatDate(value);
    return formattedDate === 'Not available' ? t('Not available') : formattedDate;
  };

  const loadCars = useCallback(async () => {
    setIsLoading(true);
    setError('');

    try {
      const loadedCars = await listAdminCars(token);
      setCars(loadedCars);
    } catch (loadError) {
      const nextMessage = loadError instanceof Error ? loadError.message : 'Unable to load admin inventory';
      setError(nextMessage);
      if (loadError instanceof ApiError && loadError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsLoading(false);
    }
  }, [onUnauthorized, token]);

  const loadCustomers = useCallback(async () => {
    setIsCustomersLoading(true);
    setError('');

    try {
      const loadedCustomers = await listAdminUsers(token);
      setCustomers(loadedCustomers);
      setRatingInputs((current) => {
        const nextInputs = { ...current };
        loadedCustomers.forEach((customer) => {
          nextInputs[customer.id] = nextInputs[customer.id] || '5';
        });
        return nextInputs;
      });
    } catch (loadError) {
      const nextMessage = loadError instanceof Error ? loadError.message : 'Unable to load customers';
      setError(nextMessage);
      if (loadError instanceof ApiError && loadError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsCustomersLoading(false);
    }
  }, [onUnauthorized, token]);

  useEffect(() => {
    loadCars();
    loadCustomers();
  }, [loadCars, loadCustomers]);

  const handleRefresh = () => {
    loadCars();
    loadCustomers();
  };

  const stats = useMemo(() => {
    const available = cars.filter((car) => car.status === 'available').length;
    const averagePrice =
      cars.length === 0 ? 0 : cars.reduce((total, car) => total + car.price_per_day, 0) / cars.length;
    const statuses = cars.reduce<Record<string, number>>((acc, car) => {
      acc[car.status] = (acc[car.status] || 0) + 1;
      return acc;
    }, {});

    return {
      total: cars.length,
      available,
      averagePrice,
      statuses,
    };
  }, [cars]);

  const updateField =
    (field: keyof CarFormState) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      setForm((current) => ({
        ...current,
        [field]: event.target.value,
      }));
    };

  const resetForm = () => {
    setForm(emptyForm);
    setEditingId(null);
  };

  const updateRatingInput =
    (userID: number) =>
    (event: ChangeEvent<HTMLSelectElement>) => {
      setRatingInputs((current) => ({
        ...current,
        [userID]: event.target.value,
      }));
    };

  const handleRateCustomer = async (customer: User) => {
    const rating = Number(ratingInputs[customer.id] || '');
    setError('');
    setMessage('');

    if (!Number.isFinite(rating) || rating < 1 || rating > 5) {
      setError('Customer rating must be between 1 and 5.');
      return;
    }

    try {
      const updatedCustomer = await rateAdminUser(token, customer.id, { rating });
      setCustomers((current) => current.map((item) => (item.id === updatedCustomer.id ? updatedCustomer : item)));
      setRatingInputs((current) => ({
        ...current,
        [customer.id]: '5',
      }));
      setMessage('Rating added.');
    } catch (rateError) {
      setError(rateError instanceof Error ? rateError.message : 'Unable to rate customer');
      if (rateError instanceof ApiError && rateError.status === 401) {
        onUnauthorized();
      }
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSaving(true);
    setError('');
    setMessage('');

    try {
      const payload = toCarInput(form);
      if (editingId) {
        await updateAdminCar(token, editingId, payload);
        setMessage('Vehicle updated.');
      } else {
        await createAdminCar(token, payload);
        setMessage('Vehicle created.');
      }

      resetForm();
      await loadCars();
      onInventoryChanged();
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : 'Unable to save vehicle');
      if (saveError instanceof ApiError && saveError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleEdit = (car: Car) => {
    setForm(formFromCar(car));
    setEditingId(car.id);
    setError('');
    setMessage('');
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleDelete = async (car: Car) => {
    const shouldDelete = window.confirm(`${t('Delete')} ${car.brand} ${car.model}?`);
    if (!shouldDelete) {
      return;
    }

    setError('');
    setMessage('');

    try {
      await deleteAdminCar(token, car.id);
      setMessage('Vehicle deleted.');
      await loadCars();
      onInventoryChanged();
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : 'Unable to delete vehicle');
      if (deleteError instanceof ApiError && deleteError.status === 401) {
        onUnauthorized();
      }
    }
  };

  return (
    <main className="pt-24 pb-16 px-4">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.22em] text-cyan-300">{t('Admin')}</p>
            <h1 className="mt-2 text-4xl font-bold text-white">{t('Fleet Dashboard')}</h1>
            <p className="mt-3 max-w-2xl text-gray-300">
              {t('Manage the live vehicle inventory served from the existing AutoRent backend.')}
            </p>
          </div>
          <button
            type="button"
            onClick={handleRefresh}
            disabled={isLoading || isCustomersLoading}
            className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-4 py-2 text-sm font-semibold text-cyan-100 hover:bg-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
          >
            <RefreshCw size={16} className={isLoading || isCustomersLoading ? 'animate-spin' : ''} />
            {t('Refresh')}
          </button>
        </div>

        <div className="inline-flex rounded-lg border border-cyan-500/20 bg-black/35 p-1">
          <button
            type="button"
            onClick={() => setActivePanel('fleet')}
            className={`inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-semibold transition-colors ${
              activePanel === 'fleet' ? 'bg-cyan-500 text-black' : 'text-cyan-100 hover:bg-cyan-500/10'
            }`}
          >
            <CarIcon size={16} />
            {t('Fleet')}
          </button>
          <button
            type="button"
            onClick={() => setActivePanel('news')}
            className={`inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-semibold transition-colors ${
              activePanel === 'news' ? 'bg-cyan-500 text-black' : 'text-cyan-100 hover:bg-cyan-500/10'
            }`}
          >
            <Newspaper size={16} />
            {t('News')}
          </button>
        </div>

        {activePanel === 'news' ? (
          <AdminNewsDashboard token={token} onNewsChanged={onNewsChanged} onUnauthorized={onUnauthorized} />
        ) : (
          <>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
            <p className="text-sm text-gray-400">{t('Total Vehicles')}</p>
            <p className="mt-2 text-3xl font-bold text-white">{stats.total}</p>
          </div>
          <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
            <p className="text-sm text-gray-400">{t('Available')}</p>
            <p className="mt-2 text-3xl font-bold text-cyan-300">{stats.available}</p>
          </div>
          <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
            <p className="text-sm text-gray-400">{t('Average Daily Price')}</p>
            <p className="mt-2 text-3xl font-bold text-white">{currencyFormatter.format(stats.averagePrice)}</p>
          </div>
        </div>

        <div className="rounded-xl border border-cyan-500/20 bg-black/30 p-5">
          <p className="mb-3 text-sm font-semibold text-gray-300">{t('Status Breakdown')}</p>
          <div className="flex flex-wrap gap-2">
            {Object.entries(stats.statuses).length === 0 ? (
              <span className="text-sm text-gray-500">{t('No vehicles yet.')}</span>
            ) : (
              Object.entries(stats.statuses).map(([status, count]) => (
                <span
                  key={status}
                  className="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-3 py-1 text-sm capitalize text-cyan-100"
                >
                  {t(status)}: {count}
                </span>
              ))
            )}
          </div>
        </div>

        <section className="rounded-xl border border-cyan-500/20 bg-white/10 p-6">
          <div className="mb-6 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Customers')}</p>
              <h2 className="mt-1 text-2xl font-semibold text-white">{t('Client ratings')}</h2>
              <p className="mt-1 text-sm text-gray-400">
                {t('Leave a rating for registered customers. New ratings update their average score.')}
              </p>
            </div>
            <span className="inline-flex items-center gap-2 rounded-full border border-cyan-500/20 bg-cyan-500/10 px-3 py-1 text-sm text-cyan-100">
              <Users size={16} />
              {customers.length} {customers.length === 1 ? t('client') : t('clients')}
            </span>
          </div>

          {isCustomersLoading ? (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3" aria-label="Loading customers">
              {[0, 1, 2].map((item) => (
                <div key={item} className="h-40 animate-pulse rounded-xl bg-black/40" />
              ))}
            </div>
          ) : customers.length === 0 ? (
            <div className="rounded-xl border border-cyan-500/10 bg-black/35 px-4 py-8 text-center text-gray-300">
              {t('No registered clients yet.')}
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {customers.map((customer) => (
                <article key={customer.id} className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h3 className="break-words text-lg font-semibold text-white">{customer.name || customer.email}</h3>
                      <p className="mt-1 break-words text-sm text-gray-400">{customer.email}</p>
                    </div>
                    <span className="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-xs capitalize text-cyan-100">
                      {t(customer.status || 'unknown')}
                    </span>
                  </div>

                  <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
                    <div className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                      <label className="block text-gray-500" htmlFor={`customer-rating-${customer.id}`}>
                        {t('Rating')}
                      </label>
                      <select
                        id={`customer-rating-${customer.id}`}
                        value={ratingInputs[customer.id] || '5'}
                        onChange={updateRatingInput(customer.id)}
                        className="mt-2 h-9 w-full rounded-lg border border-cyan-500/25 bg-black/60 px-2 text-sm font-semibold text-cyan-300 transition-colors focus:border-cyan-400 focus:outline-none"
                        aria-label={`Rating for ${customer.name || customer.email}`}
                      >
                        {ratingOptions.map((rating) => (
                          <option key={rating} value={rating} className="bg-gray-950 text-white">
                            {rating}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                      <p className="text-gray-500">{t('Count')}</p>
                      <p className="mt-1 font-semibold text-white">{customer.rating_count || 0}</p>
                    </div>
                  </div>

                  <p className="mt-3 text-xs text-gray-500">
                    {t('Joined')} {displayDate(customer.created_at)}
                  </p>

                  <div className="mt-4">
                    <button
                      type="button"
                      onClick={() => handleRateCustomer(customer)}
                      className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-3 py-2 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
                    >
                      <Star size={16} />
                      {t('Rate')}
                    </button>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>

        {(error || message) && (
          <div
            className={`rounded-xl border px-4 py-3 text-sm ${
              error ? 'border-red-400/30 bg-red-500/10 text-red-100' : 'border-cyan-400/30 bg-cyan-500/10 text-cyan-100'
            }`}
            role={error ? 'alert' : 'status'}
          >
            {t(error || message)}
          </div>
        )}

        <div className="grid grid-cols-1 gap-8 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.4fr)]">
          <form onSubmit={handleSubmit} className="rounded-xl border border-cyan-500/20 bg-white/10 p-6">
            <div className="mb-6 flex items-center justify-between gap-4">
              <div>
                <h2 className="text-2xl font-semibold text-white">{editingId ? t('Edit Vehicle') : t('Add Vehicle')}</h2>
                <p className="mt-1 text-sm text-gray-400">{t('All required fields map directly to the backend car payload.')}</p>
              </div>
              {editingId && (
                <button
                  type="button"
                  onClick={resetForm}
                  className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-500/20 text-cyan-100 hover:bg-cyan-500/10 transition-colors"
                  aria-label={t('Cancel editing')}
                >
                  <X size={18} />
                </button>
              )}
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <label className={labelClass}>
                <span>{t('Brand')}</span>
                <input value={form.brand} onChange={updateField('brand')} className={inputClass} required maxLength={50} />
              </label>
              <label className={labelClass}>
                <span>{t('Model')}</span>
                <input value={form.model} onChange={updateField('model')} className={inputClass} required maxLength={50} />
              </label>
              <label className={labelClass}>
                <span>{t('Year')}</span>
                <input type="number" value={form.year} onChange={updateField('year')} className={inputClass} required min={1900} />
              </label>
              <label className={labelClass}>
                <span>{t('Car Class')}</span>
                <input value={form.carClass} onChange={updateField('carClass')} className={inputClass} required maxLength={50} />
              </label>
              <label className={labelClass}>
                <span>{t('Body Type')}</span>
                <input value={form.bodyType} onChange={updateField('bodyType')} className={inputClass} required maxLength={50} />
              </label>
              <label className={labelClass}>
                <span>{t('Transmission')}</span>
                <input
                  value={form.transmission}
                  onChange={updateField('transmission')}
                  className={inputClass}
                  required
                  maxLength={30}
                />
              </label>
              <label className={labelClass}>
                <span>{t('Fuel Type')}</span>
                <input value={form.fuelType} onChange={updateField('fuelType')} className={inputClass} required maxLength={30} />
              </label>
              <label className={labelClass}>
                <span>{t('Status')}</span>
                <select value={form.status} onChange={updateField('status')} className={inputClass}>
                  {statusOptions.map((status) => (
                    <option key={status} value={status} className="bg-gray-950">
                      {t(status)}
                    </option>
                  ))}
                </select>
              </label>
              <label className={labelClass}>
                <span>{t('Seats')}</span>
                <input type="number" value={form.seats} onChange={updateField('seats')} className={inputClass} required min={1} />
              </label>
              <label className={labelClass}>
                <span>{t('Doors')}</span>
                <input type="number" value={form.doors} onChange={updateField('doors')} className={inputClass} required min={1} />
              </label>
              <label className={labelClass}>
                <span>{t('Engine Volume')}</span>
                <input
                  type="number"
                  step="0.1"
                  value={form.engineVolume}
                  onChange={updateField('engineVolume')}
                  className={inputClass}
                  placeholder={t('Optional')}
                />
              </label>
              <label className={labelClass}>
                <span>{t('Horsepower')}</span>
                <input
                  type="number"
                  value={form.horsepower}
                  onChange={updateField('horsepower')}
                  className={inputClass}
                  placeholder={t('Optional')}
                />
              </label>
              <label className={labelClass}>
                <span>{t('Price Per Day')}</span>
                <input
                  type="number"
                  step="0.01"
                  value={form.pricePerDay}
                  onChange={updateField('pricePerDay')}
                  className={inputClass}
                  required
                  min={0}
                />
              </label>
              <label className={labelClass}>
                <span>{t('Deposit')}</span>
                <input
                  type="number"
                  step="0.01"
                  value={form.deposit}
                  onChange={updateField('deposit')}
                  className={inputClass}
                  required
                  min={0}
                />
              </label>
              <label className={`${labelClass} md:col-span-2`}>
                <span>{t('Color')}</span>
                <input value={form.color} onChange={updateField('color')} className={inputClass} maxLength={30} placeholder={t('Optional')} />
              </label>
              <label className={`${labelClass} md:col-span-2`}>
                <span>{t('Image URLs')}</span>
                <textarea
                  value={form.imageUrls}
                  onChange={updateField('imageUrls')}
                  className={`${inputClass} min-h-28 resize-y`}
                  placeholder={t('One URL per line. The first URL becomes the main image.')}
                />
              </label>
            </div>

            <button
              type="submit"
              disabled={isSaving}
              className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-4 py-3 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
            >
              {editingId ? <Save size={17} /> : <Plus size={17} />}
              {isSaving ? t('Saving...') : editingId ? t('Save Changes') : t('Create Vehicle')}
            </button>
          </form>

          <section className="rounded-xl border border-cyan-500/20 bg-white/10 p-6">
            <div className="mb-6 flex items-center justify-between gap-4">
              <div>
                <h2 className="text-2xl font-semibold text-white">{t('Inventory')}</h2>
                <p className="mt-1 text-sm text-gray-400">{t('Create, edit, and delete cars from the admin API.')}</p>
              </div>
            </div>

            {isLoading ? (
              <div className="space-y-3" aria-label="Loading admin cars">
                {[0, 1, 2].map((item) => (
                  <div key={item} className="h-24 animate-pulse rounded-xl bg-black/40" />
                ))}
              </div>
            ) : cars.length === 0 ? (
              <div className="rounded-xl border border-cyan-500/10 bg-black/40 py-12 text-center">
                <p className="text-lg font-semibold text-white">{t('No cars in inventory.')}</p>
                <p className="mt-2 text-sm text-gray-400">{t('Use the form to add the first vehicle.')}</p>
              </div>
            ) : (
              <div className="space-y-4">
                {cars.map((car) => (
                  <article key={car.id} className="grid gap-4 rounded-xl border border-cyan-500/10 bg-black/35 p-4 lg:grid-cols-[8rem_1fr_auto]">
                    <img
                      src={mainImage(car)}
                      alt={`${car.brand} ${car.model}`}
                      referrerPolicy="no-referrer"
                      className="h-32 w-full rounded-lg object-cover lg:h-24 lg:w-32"
                      onError={(event) => {
                        event.currentTarget.onerror = null;
                        event.currentTarget.src = fallbackImageUrl;
                      }}
                    />
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <h3 className="text-lg font-semibold text-white">
                          {car.brand} {car.model}
                        </h3>
                        <span className="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-xs text-cyan-100 capitalize">
                          {displayVehicleTerm(car.status, t)}
                        </span>
                      </div>
                      <p className="mt-2 text-sm text-gray-300">
                        {car.year} | {displayVehicleTerm(car.car_class, t)} | {displayVehicleTerm(car.body_type, t)} |{' '}
                        {displayVehicleTerm(car.transmission, t)} | {displayVehicleTerm(car.fuel_type, t)}
                      </p>
                      <p className="mt-1 text-sm text-gray-400">
                        {car.seats} {t('seats')} | {car.doors} {t('doors')} | {t('Deposit')} {currencyFormatter.format(car.deposit)}
                      </p>
                      <p className="mt-2 font-semibold text-cyan-300">
                        {currencyFormatter.format(car.price_per_day)} / {t('day')}
                      </p>
                    </div>
                    <div className="flex items-center gap-2 lg:flex-col lg:items-stretch">
                      <button
                        type="button"
                        onClick={() => handleEdit(car)}
                        className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 hover:bg-cyan-500/10 transition-colors"
                      >
                        <Edit3 size={16} />
                        {t('Edit')}
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(car)}
                        className="inline-flex items-center justify-center gap-2 rounded-lg border border-red-400/30 px-3 py-2 text-sm font-semibold text-red-200 hover:bg-red-500/10 transition-colors"
                      >
                        <Trash2 size={16} />
                        {t('Delete')}
                      </button>
                    </div>
                  </article>
                ))}
              </div>
            )}
          </section>
        </div>
          </>
        )}
      </div>
    </main>
  );
};

export default AdminDashboard;
