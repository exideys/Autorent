import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Car } from '../../types/api';
import { renderWithTranslations } from '../../test/render';
import VehicleDetailsModal from '../VehicleDetailsModal';

const car: Car = {
  id: 11,
  brand: 'Mercedes',
  model: 'S-Class',
  year: 2025,
  car_class: 'Luxury',
  body_type: 'Sedan',
  transmission: 'Automatic',
  fuel_type: 'Hybrid',
  seats: 5,
  doors: 4,
  engine_volume: 3,
  horsepower: 430,
  price_per_day: 400,
  deposit: 1500,
  color: 'Black',
  status: 'available',
  created_at: '2026-01-01T00:00:00Z',
  images: [
    { id: 1, car_id: 11, image_url: '/first.jpg', is_main: true, sort_order: 0 },
    { id: 2, car_id: 11, image_url: '/second.jpg', is_main: false, sort_order: 1 },
  ],
};

describe('VehicleDetailsModal', () => {
  it('opens a fullscreen gallery and switches images with arrows', async () => {
    const user = userEvent.setup();

    renderWithTranslations(<VehicleDetailsModal car={car} onBooking={vi.fn()} onClose={vi.fn()} />);

    await user.click(screen.getByRole('button', { name: /open vehicle image gallery/i }));

    expect(screen.getByRole('dialog', { name: /vehicle image gallery/i })).toBeInTheDocument();
    expect(screen.getByAltText(/image 1 of 2/i)).toHaveAttribute('src', '/first.jpg');

    await user.click(screen.getByRole('button', { name: /next image/i }));

    expect(screen.getByAltText(/image 2 of 2/i)).toHaveAttribute('src', '/second.jpg');

    await user.click(screen.getByRole('button', { name: /previous image/i }));

    expect(screen.getByAltText(/image 1 of 2/i)).toHaveAttribute('src', '/first.jpg');
  });
});
