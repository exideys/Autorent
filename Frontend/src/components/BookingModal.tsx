import { motion } from 'framer-motion';
import { Calendar, Clock, MapPin, Phone, X } from 'lucide-react';
import { useState } from 'react';
import { currencyFormatter } from '../lib/carDisplay';
import { createRentalOrder } from '../lib/api';
import type { Car } from '../types/api';

interface BookingModalProps {
  car: Car;
  token?: string;
  onCreated?: () => void | Promise<void>;
  onClose: () => void;
}

interface BookingFormState {
  phone: string;
  pickupLocation: string;
  pickupDate: string;
  returnDate: string;
  pickupTime: string;
  notes: string;
}

const initialForm: BookingFormState = {
  phone: '',
  pickupLocation: '',
  pickupDate: '',
  returnDate: '',
  pickupTime: '',
  notes: '',
};

const fieldClass =
  'h-11 w-full rounded-lg border border-cyan-500/25 bg-black/60 px-3 text-sm text-white placeholder-gray-500 transition-colors focus:border-cyan-300 focus:outline-none';

const BookingModal = ({ car, token, onCreated, onClose }: BookingModalProps) => {
  const [form, setForm] = useState<BookingFormState>(initialForm);
  const [error, setError] = useState('');
  const [isCreated, setIsCreated] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const updateField = <K extends keyof BookingFormState>(field: K, value: BookingFormState[K]) => {
    setForm((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const handleSubmit = async () => {
    setError('');
    setIsCreated(false);

    if (!token) {
      setError('Please sign in to book a vehicle.');
      return;
    }

    const requiredFields = [
      form.phone.trim(),
      form.pickupLocation.trim(),
      form.pickupDate,
      form.returnDate,
      form.pickupTime,
    ];

    if (requiredFields.some((value) => !value)) {
      setError('Please fill in all booking fields.');
      return;
    }

    if (new Date(form.returnDate) < new Date(form.pickupDate)) {
      setError('Return date cannot be before pickup date.');
      return;
    }

    setIsSubmitting(true);
    try {
      await createRentalOrder(token, {
        car_id: car.id,
        start_date: form.pickupDate,
        end_date: form.returnDate,
        pickup_location: form.pickupLocation.trim(),
        pickup_time: form.pickupTime,
        phone: form.phone.trim(),
        notes: form.notes.trim() || undefined,
      });
      setIsCreated(true);
      await onCreated?.();
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Unable to create booking.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[80] flex items-center justify-center bg-black/75 px-4 py-6 backdrop-blur-sm">
      <motion.div
        initial={{ opacity: 0, y: 24, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        className="max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl border border-cyan-500/30 bg-gray-950 shadow-2xl"
      >
        <div className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-cyan-500/20 bg-gray-950/95 p-5 backdrop-blur">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">Booking</p>
            <h3 className="mt-1 text-2xl font-bold text-white">
              {car.brand} {car.model}
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-100 transition-colors hover:bg-cyan-500/10"
            aria-label="Close booking"
          >
            <X size={20} />
          </button>
        </div>

        <div className="space-y-5 p-5">
          <div className="grid gap-3 rounded-xl border border-cyan-500/10 bg-black/35 p-4 sm:grid-cols-2">
            <div>
              <p className="text-sm text-gray-400">Daily price</p>
              <p className="mt-1 text-xl font-bold text-cyan-300">{currencyFormatter.format(car.price_per_day)}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">Deposit</p>
              <p className="mt-1 text-xl font-bold text-white">{currencyFormatter.format(car.deposit)}</p>
            </div>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <label className="block space-y-2 text-sm text-gray-300">
              <span className="flex items-center gap-2">
                <Phone size={16} className="text-cyan-300" />
                Phone
              </span>
              <input value={form.phone} onChange={(event) => updateField('phone', event.target.value)} className={fieldClass} placeholder="+1 555 123 4567" />
            </label>
            <label className="block space-y-2 text-sm text-gray-300">
              <span className="flex items-center gap-2">
                <MapPin size={16} className="text-cyan-300" />
                Pickup location
              </span>
              <input
                value={form.pickupLocation}
                onChange={(event) => updateField('pickupLocation', event.target.value)}
                className={fieldClass}
                placeholder="Hotel, airport, office"
              />
            </label>
            <label className="block space-y-2 text-sm text-gray-300">
              <span className="flex items-center gap-2">
                <Calendar size={16} className="text-cyan-300" />
                Pickup date
              </span>
              <input type="date" value={form.pickupDate} onChange={(event) => updateField('pickupDate', event.target.value)} className={fieldClass} />
            </label>
            <label className="block space-y-2 text-sm text-gray-300">
              <span className="flex items-center gap-2">
                <Calendar size={16} className="text-cyan-300" />
                Return date
              </span>
              <input type="date" value={form.returnDate} onChange={(event) => updateField('returnDate', event.target.value)} className={fieldClass} />
            </label>
            <label className="block space-y-2 text-sm text-gray-300">
              <span className="flex items-center gap-2">
                <Clock size={16} className="text-cyan-300" />
                Pickup time
              </span>
              <input type="time" value={form.pickupTime} onChange={(event) => updateField('pickupTime', event.target.value)} className={fieldClass} />
            </label>
            <label className="block space-y-2 text-sm text-gray-300 sm:col-span-2">
              <span>Notes</span>
              <textarea
                value={form.notes}
                onChange={(event) => updateField('notes', event.target.value)}
                className="min-h-24 w-full resize-y rounded-lg border border-cyan-500/25 bg-black/60 px-3 py-3 text-sm text-white placeholder-gray-500 transition-colors focus:border-cyan-300 focus:outline-none"
                placeholder="Delivery details, child seat, route notes"
              />
            </label>
          </div>

          {error && (
            <p className="rounded-lg border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200" role="alert">
              {error}
            </p>
          )}
          {isCreated && (
            <p className="rounded-lg border border-cyan-400/30 bg-cyan-500/10 px-3 py-2 text-sm text-cyan-100">
              Booking created. You can view it in your profile.
            </p>
          )}

          <button
            type="button"
            onClick={handleSubmit}
            disabled={isSubmitting || isCreated}
            className="w-full rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-5 py-3 font-semibold text-white shadow-lg shadow-cyan-500/20 transition-colors hover:from-cyan-600 hover:to-violet-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isSubmitting ? 'Creating Booking...' : isCreated ? 'Booking Created' : 'Create Booking'}
          </button>
        </div>
      </motion.div>
    </div>
  );
};

export default BookingModal;
