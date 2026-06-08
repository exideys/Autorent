import { motion } from 'framer-motion';
import { Calendar, Clock, MapPin } from 'lucide-react';
import { useRef } from 'react';
import type { BookingFormErrors, BookingFormValues, CarCategory } from '../types/site';

interface BookingFormSectionProps {
  values: BookingFormValues;
  errors: BookingFormErrors;
  categoryOptions: Array<CarCategory | 'Any'>;
  onValueChange: <K extends keyof BookingFormValues>(field: K, value: BookingFormValues[K]) => void;
  onSubmit: () => void;
}

const dateInputValue = (date: Date) => {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, '0');
  const day = `${date.getDate()}`.padStart(2, '0');
  return `${year}-${month}-${day}`;
};

const timeInputValue = (date: Date) => {
  const hours = `${date.getHours()}`.padStart(2, '0');
  const minutes = `${date.getMinutes()}`.padStart(2, '0');
  return `${hours}:${minutes}`;
};

const openNativePicker = (input: HTMLInputElement | null) => {
  input?.focus();
  try {
    input?.showPicker?.();
  } catch {
    // Some browsers only allow showPicker during direct input activation.
  }
};

const BookingFormSection = ({
  values,
  errors,
  categoryOptions,
  onValueChange,
  onSubmit,
}: BookingFormSectionProps) => {
  const pickupDateRef = useRef<HTMLInputElement>(null);
  const returnDateRef = useRef<HTMLInputElement>(null);
  const pickupTimeRef = useRef<HTMLInputElement>(null);
  const now = new Date();
  const today = dateInputValue(now);
  const currentTime = timeInputValue(now);
  const minimumReturnDate = values.pickupDate && values.pickupDate > today ? values.pickupDate : today;
  const minimumPickupTime = values.pickupDate === today ? currentTime : undefined;

  return (
    <section className="py-16 px-4">
      <div className="max-w-5xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 50 }}
          whileInView={{ opacity: 1, y: 0 }}
          className="bg-white/10 backdrop-blur-lg rounded-3xl p-8 border border-cyan-500/20 shadow-2xl"
        >
          <h2 className="text-3xl font-bold text-center mb-8 text-cyan-400">Book Your Ride</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
            <div className="space-y-2">
              <label htmlFor="pickup-location" className="text-sm text-gray-300 flex items-center">
                <MapPin size={16} className="mr-2 text-cyan-400" />
                Pickup Location
              </label>
              <input
                id="pickup-location"
                type="text"
                value={values.location}
                onChange={(event) => onValueChange('location', event.target.value)}
                placeholder="Enter location"
                aria-invalid={Boolean(errors.location)}
                aria-describedby={errors.location ? 'pickup-location-error' : undefined}
                className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white placeholder-gray-500 focus:border-cyan-400 focus:outline-none transition-colors"
              />
              {errors.location && (
                <p id="pickup-location-error" className="text-xs text-red-300">
                  {errors.location}
                </p>
              )}
            </div>

            <div className="space-y-2" onClick={() => openNativePicker(pickupDateRef.current)}>
              <label htmlFor="pickup-date" className="text-sm text-gray-300 flex items-center">
                <Calendar size={16} className="mr-2 text-cyan-400" />
                Pickup Date
              </label>
              <input
                ref={pickupDateRef}
                id="pickup-date"
                type="date"
                min={today}
                value={values.pickupDate}
                onClick={() => openNativePicker(pickupDateRef.current)}
                onChange={(event) => onValueChange('pickupDate', event.target.value)}
                aria-invalid={Boolean(errors.pickupDate)}
                aria-describedby={errors.pickupDate ? 'pickup-date-error' : undefined}
                className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
              />
              {errors.pickupDate && (
                <p id="pickup-date-error" className="text-xs text-red-300">
                  {errors.pickupDate}
                </p>
              )}
            </div>

            <div className="space-y-2" onClick={() => openNativePicker(returnDateRef.current)}>
              <label htmlFor="return-date" className="text-sm text-gray-300 flex items-center">
                <Calendar size={16} className="mr-2 text-cyan-400" />
                Return Date
              </label>
              <input
                ref={returnDateRef}
                id="return-date"
                type="date"
                min={minimumReturnDate}
                value={values.returnDate}
                onClick={() => openNativePicker(returnDateRef.current)}
                onChange={(event) => onValueChange('returnDate', event.target.value)}
                aria-invalid={Boolean(errors.returnDate)}
                aria-describedby={errors.returnDate ? 'return-date-error' : undefined}
                className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
              />
              {errors.returnDate && (
                <p id="return-date-error" className="text-xs text-red-300">
                  {errors.returnDate}
                </p>
              )}
            </div>

            <div className="space-y-2" onClick={() => openNativePicker(pickupTimeRef.current)}>
              <label htmlFor="pickup-time" className="text-sm text-gray-300 flex items-center">
                <Clock size={16} className="mr-2 text-cyan-400" />
                Pickup Time
              </label>
              <input
                ref={pickupTimeRef}
                id="pickup-time"
                type="time"
                min={minimumPickupTime}
                value={values.pickupTime}
                onClick={() => openNativePicker(pickupTimeRef.current)}
                onChange={(event) => onValueChange('pickupTime', event.target.value)}
                aria-invalid={Boolean(errors.pickupTime)}
                aria-describedby={errors.pickupTime ? 'pickup-time-error' : undefined}
                className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
              />
              {errors.pickupTime && (
                <p id="pickup-time-error" className="text-xs text-red-300">
                  {errors.pickupTime}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <label htmlFor="car-category" className="text-sm text-gray-300">
                Vehicle Class
              </label>
              <select
                id="car-category"
                value={values.category}
                onChange={(event) => onValueChange('category', event.target.value as BookingFormValues['category'])}
                className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
              >
                {categoryOptions.map((category) => (
                  <option key={category} value={category} className="bg-gray-900">
                    {category === 'Any' ? 'Any class' : category}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div className="text-center mt-8">
            <button
              type="button"
              onClick={onSubmit}
              className="bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 text-white font-semibold py-3 px-8 rounded-3xl transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-cyan-500/25"
            >
              Search Vehicles
            </button>
          </div>
        </motion.div>
      </div>
    </section>
  );
};

export default BookingFormSection;
