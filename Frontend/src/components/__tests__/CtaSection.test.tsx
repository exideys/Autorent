import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithTranslations } from '../../test/render';
import CtaSection from '../CtaSection';

describe('CtaSection', () => {
  it('calls the showroom navigation handler from the get started button', async () => {
    const user = userEvent.setup();
    const onGetStarted = vi.fn();

    renderWithTranslations(<CtaSection onGetStarted={onGetStarted} />);

    await user.click(screen.getByRole('button', { name: /get started today/i }));

    expect(onGetStarted).toHaveBeenCalledTimes(1);
  });
});
