import { motion } from 'framer-motion';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { resolveApiAssetUrl } from '../lib/api';
import { currencyFormatter, detailRows, displayVehicleTerm, fallbackImageUrl, translatableVehicleTerms } from '../lib/carDisplay';
import type { Car } from '../types/api';

interface VehicleDetailsModalProps {
  car: Car;
  onBooking: (car: Car) => void;
  onClose: () => void;
}

const VehicleDetailsModal = ({ car, onBooking, onClose }: VehicleDetailsModalProps) => {
  const galleryImages = useMemo(() => {
    const resolvedImages = car.images.map((image) => ({
      id: image.id,
      url: resolveApiAssetUrl(image.image_url, fallbackImageUrl),
      isMain: image.is_main,
    }));

    return resolvedImages.length > 0 ? resolvedImages : [{ id: 0, url: fallbackImageUrl, isMain: true }];
  }, [car.images]);
  const defaultImageIndex = useMemo(() => {
    const mainIndex = galleryImages.findIndex((image) => image.isMain);
    return mainIndex >= 0 ? mainIndex : 0;
  }, [galleryImages]);
  const [selectedImageIndex, setSelectedImageIndex] = useState(defaultImageIndex);
  const [isGalleryOpen, setIsGalleryOpen] = useState(false);
  const selectedImage = galleryImages[selectedImageIndex] || galleryImages[0];
  const thumbnailImages = galleryImages.slice(0, 6);
  const rows = detailRows(car);
  const translatableValueLabels = new Set(['Class', 'Body type', 'Transmission', 'Fuel type', 'Color', 'Status']);
  const { t } = useTranslation([
    'Vehicle details',
    'Daily price',
    'Deposit',
    'Booking',
    'Not specified',
    'Vehicle image gallery',
    'Open vehicle image gallery',
    'Previous image',
    'Next image',
    'Close image gallery',
    ...rows.map(([label]) => String(label)),
    ...translatableVehicleTerms(
      rows.filter(([label, value]) => translatableValueLabels.has(String(label)) && typeof value === 'string').map(([, value]) => String(value)),
    ),
  ]);

  const detailValue = (label: string, value: string | number) => {
    if (typeof value === 'string' && (translatableValueLabels.has(label) || value === 'Not specified')) {
      return displayVehicleTerm(value, t);
    }

    return value;
  };

  useEffect(() => {
    setSelectedImageIndex(defaultImageIndex);
  }, [defaultImageIndex]);

  const showPreviousImage = () => {
    setSelectedImageIndex((current) => (current === 0 ? galleryImages.length - 1 : current - 1));
  };

  const showNextImage = () => {
    setSelectedImageIndex((current) => (current + 1) % galleryImages.length);
  };

  useEffect(() => {
    if (!isGalleryOpen) {
      return undefined;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsGalleryOpen(false);
      }
      if (event.key === 'ArrowLeft') {
        showPreviousImage();
      }
      if (event.key === 'ArrowRight') {
        showNextImage();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isGalleryOpen, galleryImages.length]);

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/75 px-4 py-6 backdrop-blur-sm">
      <motion.div
        initial={{ opacity: 0, y: 24, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        className="max-h-[90vh] w-full max-w-5xl overflow-y-auto rounded-xl border border-cyan-500/30 bg-gray-950 shadow-2xl"
      >
        <div className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-cyan-500/20 bg-gray-950/95 p-5 backdrop-blur">
          <div>
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Vehicle details')}</p>
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
            <button
              type="button"
              onClick={() => setIsGalleryOpen(true)}
              className="block w-full overflow-hidden rounded-xl border border-cyan-500/15 text-left transition-colors hover:border-cyan-300/60 focus:outline-none focus:ring-2 focus:ring-cyan-300/60"
              aria-label={t('Open vehicle image gallery')}
            >
              <img
                src={selectedImage?.url || fallbackImageUrl}
                alt={`${car.brand} ${car.model}`}
                referrerPolicy="no-referrer"
                className="h-72 w-full object-cover md:h-80"
                onError={(event) => {
                  event.currentTarget.onerror = null;
                  event.currentTarget.src = fallbackImageUrl;
                }}
              />
            </button>
            {thumbnailImages.length > 1 && (
              <div className="grid grid-cols-3 gap-3">
                {thumbnailImages.map((image, index) => {
                  const isSelected = index === selectedImageIndex;

                  return (
                    <button
                      key={image.id}
                      type="button"
                      onClick={() => setSelectedImageIndex(index)}
                      className={`overflow-hidden rounded-lg border transition-all ${
                        isSelected
                          ? 'border-cyan-300 shadow-lg shadow-cyan-500/25'
                          : 'border-cyan-500/20 hover:border-cyan-300/60'
                      }`}
                      aria-label={`Show ${car.brand} ${car.model} image`}
                      aria-pressed={isSelected}
                    >
                      <img
                        src={image.url}
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
                <p className="text-sm text-gray-400">{t('Daily price')}</p>
                <p className="mt-1 text-2xl font-bold text-cyan-300">{currencyFormatter.format(car.price_per_day)}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/20 bg-black/35 p-4">
                <p className="text-sm text-gray-400">{t('Deposit')}</p>
                <p className="mt-1 text-2xl font-bold text-white">{currencyFormatter.format(car.deposit)}</p>
              </div>
            </div>

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {rows.map(([label, value]) => (
                <div key={label} className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                  <p className="text-xs uppercase tracking-wide text-gray-500">{t(String(label))}</p>
                  <p className="mt-1 text-sm font-semibold capitalize text-gray-100">{detailValue(String(label), value)}</p>
                </div>
              ))}
            </div>

            <button
              type="button"
              onClick={() => onBooking(car)}
              className="w-full rounded-lg bg-gradient-to-r from-cyan-500 to-violet-600 px-5 py-3 font-semibold text-white shadow-lg shadow-cyan-500/20 transition-colors hover:from-cyan-600 hover:to-violet-700"
            >
              {t('Booking')}
            </button>
          </div>
        </div>
      </motion.div>

      {isGalleryOpen && (
        <div
          className="fixed inset-0 z-[90] flex items-center justify-center bg-black/95 px-3 py-4 backdrop-blur-md sm:px-6"
          role="dialog"
          aria-modal="true"
          aria-label={t('Vehicle image gallery')}
        >
          <button
            type="button"
            onClick={() => setIsGalleryOpen(false)}
            className="absolute right-4 top-4 inline-flex h-11 w-11 items-center justify-center rounded-lg border border-white/20 bg-black/50 text-white transition-colors hover:bg-white/10"
            aria-label={t('Close image gallery')}
          >
            <X size={22} />
          </button>

          {galleryImages.length > 1 && (
            <button
              type="button"
              onClick={showPreviousImage}
              className="absolute left-3 top-1/2 inline-flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full border border-white/20 bg-black/55 text-white transition-colors hover:bg-white/10 sm:left-6"
              aria-label={t('Previous image')}
            >
              <ChevronLeft size={28} />
            </button>
          )}

          <img
            src={selectedImage?.url || fallbackImageUrl}
            alt={`${car.brand} ${car.model} image ${selectedImageIndex + 1} of ${galleryImages.length}`}
            referrerPolicy="no-referrer"
            className="max-h-[88vh] max-w-[92vw] rounded-lg object-contain shadow-2xl shadow-black"
            onError={(event) => {
              event.currentTarget.onerror = null;
              event.currentTarget.src = fallbackImageUrl;
            }}
          />

          {galleryImages.length > 1 && (
            <button
              type="button"
              onClick={showNextImage}
              className="absolute right-3 top-1/2 inline-flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full border border-white/20 bg-black/55 text-white transition-colors hover:bg-white/10 sm:right-6"
              aria-label={t('Next image')}
            >
              <ChevronRight size={28} />
            </button>
          )}
        </div>
      )}
    </div>
  );
};

export default VehicleDetailsModal;
