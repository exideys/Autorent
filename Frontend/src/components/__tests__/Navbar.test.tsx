import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { PageLink } from '../../types/site';
import { renderWithTranslations } from '../../test/render';
import Navbar from '../Navbar';

const pages: PageLink[] = [
  { key: 'home', label: 'Home' },
  { key: 'services', label: 'Services' },
  { key: 'showroom', label: 'Showroom' },
  { key: 'news', label: 'News' },
  { key: 'how-it-works', label: 'How It Works' },
  { key: 'contact', label: 'Contact' },
];

describe('Navbar', () => {
  it('renders every desktop navigation item', () => {
    renderWithTranslations(
      <Navbar pages={pages} activePage="home" isMenuOpen={false} onToggleMenu={vi.fn()} onNavigate={vi.fn()} />,
    );

    pages.forEach((page) => {
      expect(screen.getByRole('button', { name: page.label })).toBeInTheDocument();
    });
  });

  it('navigates to the clicked page', async () => {
    const user = userEvent.setup();
    const onNavigate = vi.fn();

    renderWithTranslations(
      <Navbar pages={pages} activePage="home" isMenuOpen={false} onToggleMenu={vi.fn()} onNavigate={onNavigate} />,
    );

    await user.click(screen.getByRole('button', { name: 'Showroom' }));

    expect(onNavigate).toHaveBeenCalledWith('showroom');
  });

  it('keeps the mobile menu toggle available below desktop layouts', async () => {
    const user = userEvent.setup();
    const onToggleMenu = vi.fn();

    renderWithTranslations(
      <Navbar pages={pages} activePage="home" isMenuOpen={false} onToggleMenu={onToggleMenu} onNavigate={vi.fn()} />,
    );

    const toggleButton = screen.getByRole('button', { name: /open navigation menu/i });

    expect(toggleButton).toHaveAttribute('aria-controls', 'mobile-navigation');

    await user.click(toggleButton);

    expect(onToggleMenu).toHaveBeenCalledTimes(1);
  });
});
