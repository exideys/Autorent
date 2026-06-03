import type { Car } from '../types/api';
import { resolveApiAssetUrl } from './api';

export const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

export const fallbackImageUrl = `${import.meta.env.BASE_URL}hero-main.png`;

export const mainImage = (car: Car) => {
  const selectedImage = car.images.find((image) => image.is_main) || car.images[0];
  return resolveApiAssetUrl(selectedImage?.image_url, fallbackImageUrl);
};

const fixedVehicleTerms = new Set(['SUV']);

export const isFixedVehicleTerm = (value?: string | null) => fixedVehicleTerms.has((value || '').trim().toUpperCase());

export const displayVehicleTerm = (value: string, translate: (text: string) => string) =>
  isFixedVehicleTerm(value) ? value.trim().toUpperCase() : translate(value);

export const translatableVehicleTerms = (values: ReadonlyArray<string | undefined | null>) =>
  values.filter((value): value is string => Boolean(value && !isFixedVehicleTerm(value)));

export const detailRows = (car: Car) => [
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
