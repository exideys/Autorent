import { motion } from 'framer-motion';
import { ArrowRight, Calendar, Gauge, Shield, Users } from 'lucide-react';
import { useState } from 'react';
import { currencyFormatter, fallbackImageUrl, mainImage } from '../lib/carDisplay';
import type { Car } from '../types/api';
import BookingModal from './BookingModal';
import VehicleDetailsModal from './VehicleDetailsModal';

interface PopularCarsSectionProps {
  cars: Car[];
  error: string;
  isLoading: boolean;
  onExploreShowroom: () => void;
}

const PopularCarsSection = ({ cars, error, isLoading, onExploreShowroom }: PopularCarsSectionProps) => {
  const [selectedCar, setSelectedCar] = useState<Car | null>(null);
  const [bookingCar, setBookingCar] = useState<Car | null>(null);
  const popularCars = cars.slice(0, 10);

  return (
    <section className="px-4 py-16">
      <div className="mx-auto max-w-7xl">
        <div className="mb-10 flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.24em] text-cyan-300">Fleet favorites</p>
            <h2 className="mt-2 text-4xl font-bold text-white">Most Popular Cars</h2>
            <p className="mt-3 max-w-2xl text-gray-300">Customer favorites from our live fleet, curated for a quick first look.</p>
          </div>
          <button
            type="button"
            onClick={onExploreShowroom}
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-cyan-500 px-5 py-3 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
          >
            Explore Showroom
            <ArrowRight size={17} />
          </button>
        </div>

        {error ? (
          <div className="rounded-xl border border-cyan-500/10 bg-black/35 px-5 py-8 text-center text-gray-300">
            Popular cars are temporarily unavailable.
          </div>
        ) : isLoading ? (
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-5" aria-label="Loading popular cars">
            {[0, 1, 2, 3, 4].map((item) => (
              <div key={item} className="h-80 animate-pulse rounded-xl border border-cyan-500/10 bg-black/40" />
            ))}
          </div>
        ) : popularCars.length === 0 ? (
          <div className="rounded-xl border border-cyan-500/10 bg-black/35 px-5 py-8 text-center text-gray-300">
            No popular cars are available yet.
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-5">
            {popularCars.map((car, index) => (
              <motion.article
                key={car.id}
                initial={{ opacity: 0, y: 24 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.04 }}
                className="overflow-hidden rounded-xl border border-cyan-500/20 bg-black/45 shadow-lg shadow-black/20"
              >
                <div className="relative h-40">
                  <img
                    src={mainImage(car)}
                    alt={`${car.brand} ${car.model}`}
                    className="h-full w-full object-cover"
                    onError={(event) => {
                      event.currentTarget.onerror = null;
                      event.currentTarget.src = fallbackImageUrl;
                    }}
                  />
                  <div className="absolute inset-0 bg-gradient-to-t from-black/85 to-transparent" />
                  <span className="absolute left-3 top-3 rounded-full border border-cyan-300/30 bg-cyan-500/20 px-3 py-1 text-xs text-cyan-100">
                    #{index + 1}
                  </span>
                </div>
                <div className="p-4">
                  <h3 className="min-h-14 text-lg font-semibold text-white">
                    {car.brand} {car.model}
                  </h3>
                  <div className="mt-3 space-y-2 text-sm text-gray-300">
                    <div className="flex items-center gap-2">
                      <Calendar size={15} className="text-cyan-300" />
                      <span>
                        {car.year} | {car.body_type}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Users size={15} className="text-cyan-300" />
                      <span>{car.seats} seats</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Gauge size={15} className="text-cyan-300" />
                      <span>{car.transmission}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Shield size={15} className="text-cyan-300" />
                      <span>{car.car_class}</span>
                    </div>
                  </div>
                  <p className="mt-4 font-semibold text-cyan-300">{currencyFormatter.format(car.price_per_day)} / day</p>
                  <div className="mt-4 grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => setSelectedCar(car)}
                      className="rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                    >
                      Details
                    </button>
                    <button
                      type="button"
                      onClick={() => setBookingCar(car)}
                      className="rounded-lg bg-cyan-500 px-3 py-2 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
                    >
                      Booking
                    </button>
                  </div>
                </div>
              </motion.article>
            ))}
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

export default PopularCarsSection;
