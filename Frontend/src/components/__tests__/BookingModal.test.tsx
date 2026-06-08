import { fireEvent, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Car } from '../../types/api';
import { renderWithTranslations } from '../../test/render';
import BookingModal from '../BookingModal';
import { createRentalOrder } from '../../lib/api';

vi.mock('../../lib/api', () => ({
  createRentalOrder: vi.fn(),
  translateTexts: vi.fn(),
}));

const car: Car = {
  id: 7,
  brand: 'BMW',
  model: 'X5',
  year: 2024,
  car_class: 'Luxury',
  body_type: 'SUV',
  transmission: 'Automatic',
  fuel_type: 'Petrol',
  seats: 5,
  doors: 4,
  price_per_day: 250,
  deposit: 1000,
  status: 'available',
  created_at: '2026-01-01T00:00:00Z',
  images: [],
};

describe('BookingModal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('rejects a pickup date earlier than today before creating an order', async () => {
    const user = userEvent.setup();

    renderWithTranslations(<BookingModal car={car} token="token" onClose={vi.fn()} />);

    fireEvent.change(screen.getByLabelText(/phone/i), { target: { value: '+1 555 123 4567' } });
    fireEvent.change(screen.getByLabelText(/pickup location/i), { target: { value: 'Airport' } });
    fireEvent.change(screen.getByLabelText(/pickup date/i), { target: { value: '2000-01-01' } });
    fireEvent.change(screen.getByLabelText(/return date/i), { target: { value: '2000-01-02' } });
    fireEvent.change(screen.getByLabelText(/pickup time/i), { target: { value: '09:30' } });

    await user.click(screen.getByRole('button', { name: /create booking/i }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Pickup date cannot be earlier than today.');
    expect(createRentalOrder).not.toHaveBeenCalled();
  });
});
