import { motion } from 'framer-motion';
import { X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { currencyFormatter, detailRows, fallbackImageUrl, mainImage } from '../lib/carDisplay';
import type { Car } from '../types/api';

interface VehicleDetailsModalProps {
  car: Car;
  onBooking: (car: Car) => void;
  onClose: () => void;
}

const VehicleDetailsModal = ({ car, onBooking, onClose }: VehicleDetailsModalProps) => {
  const defaultImageUrl = useMemo(() => mainImage(car), [car]);
  const [selectedImageUrl, setSelectedImageUrl] = useState(defaultImageUrl);
  const thumbnailImages = car.images.slice(0, 6);

  useEffect(() => {
    setSelectedImageUrl(defaultImageUrl);
  }, [defaultImageUrl]);

  return (
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
              {car.brand} {car.model}
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-100 transition-colors hover:bg-cyan-500/10"
            aria-label="Close vehicle details"
          >
            <X size={20} />
          </button>
        </div>

        <div className="grid gap-6 p-5 lg:grid-cols-[minmax(0,1.05fr)_minmax(20rem,0.95fr)]">
          <div className="space-y-4">
            <img
              src={selectedImageUrl || fallbackImageUrl}
              alt={`${car.brand} ${car.model}`}
              referrerPolicy="no-referrer"
              className="h-72 w-full rounded-xl object-cover md:h-80"
              onError={(event) => {
                event.currentTarget.onerror = null;
                event.currentTarget.src = fallbackImageUrl;
              }}
            />
            {thumbnailImages.length > 1 && (
              <div className="grid grid-cols-3 gap-3">
                {thumbnailImages.map((image) => {
                  const isSelected = image.image_url === selectedImageUrl;

                  return (
                    <button
                      key={image.id}
                      type="button"
                      onClick={() => setSelectedImageUrl(image.image_url)}
                      className={`overflow-hidden rounded-lg border transition-all ${
                        isSelected
                          ? 'border-cyan-300 shadow-lg shadow-cyan-500/25'
                          : 'border-cyan-500/20 hover:border-cyan-300/60'
                      }`}
                      aria-label={`Show ${car.brand} ${car.model} image`}
                      aria-pressed={isSelected}
                    >
                      <img
                        src={image.image_url}
                        alt={`${car.brand} ${car.model}`}
                        referrerPolicy="no-referrer"
                        className="h-24 w-full object-cover"
                        onError={(event) => {
                          event.currentTarget.style.opacity = '0.35';
                        }}
                      />
                    </button>
                  );
                })}
              </div>
            )}
          </div>

          <div className="space-y-5">
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-xl border border-cyan-500/20 bg-black/35 p-4">
                <p className="text-sm text-gray-400">Daily price</p>
                <p className="mt-1 text-2xl font-bold text-cyan-300">{currencyFormatter.format(car.price_per_day)}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/20 bg-black/35 p-4">
                <p className="text-sm text-gray-400">Deposit</p>
                <p className="mt-1 text-2xl font-bold text-white">{currencyFormatter.format(car.deposit)}</p>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {detailRows(car).map(([label, value]) => (
                <div key={label} className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                  <p className="text-xs uppercase tracking-wide text-gray-500">{label}</p>
                  <p className="mt-1 text-sm font-semibold capitalize text-gray-100">{value}</p>
                </div>
              ))}
            </div>

            <button
              type="button"
              onClick={() => onBooking(car)}
              className="w-full rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-5 py-3 font-semibold text-white shadow-lg shadow-cyan-500/20 transition-colors hover:from-cyan-600 hover:to-violet-700"
            >
              Booking
            </button>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default VehicleDetailsModal;
