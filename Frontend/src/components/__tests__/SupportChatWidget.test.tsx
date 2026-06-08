import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { User } from '../../types/api';
import { renderWithTranslations } from '../../test/render';
import SupportChatWidget from '../SupportChatWidget';
import { getSupportConversation, sendSupportMessage, streamSupportEvents } from '../../lib/api';

vi.mock('../../lib/api', () => ({
  ApiError: class ApiError extends Error {
    status: number;

    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
  downloadSupportAttachment: vi.fn(),
  getSupportConversation: vi.fn(),
  sendSupportMessage: vi.fn(),
  streamSupportEvents: vi.fn(),
  translateTexts: vi.fn(),
}));

const user: User = {
  id: 1,
  first_name: 'Jane',
  last_name: 'Driver',
  name: 'Jane Driver',
  email: 'jane@example.com',
  rating: 5,
  rating_count: 1,
  role: 'user',
  status: 'active',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
};

describe('SupportChatWidget', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getSupportConversation).mockResolvedValue({
      id: 1,
      user_id: user.id,
      status: 'open',
      messages: [],
    });
    vi.mocked(sendSupportMessage).mockResolvedValue({
      id: 1,
      conversation_id: 1,
      sender_id: user.id,
      sender_role: 'user',
      body: 'Hello',
      created_at: '2026-01-01T00:00:00Z',
      attachments: [],
    });
    vi.mocked(streamSupportEvents).mockResolvedValue(undefined);
  });

  it('sends a typed message when Enter is pressed', async () => {
    const driver = userEvent.setup();

    renderWithTranslations(<SupportChatWidget token="token" user={user} onUnauthorized={vi.fn()} />);

    await driver.click(screen.getByRole('button', { name: /open support chat/i }));
    const input = await screen.findByPlaceholderText(/write a message/i);

    await driver.type(input, 'Hello');
    await driver.keyboard('{Enter}');

    await waitFor(() => {
      expect(sendSupportMessage).toHaveBeenCalledWith('token', 'Hello', []);
    });
  });

  it('keeps a newline for Shift Enter without sending immediately', async () => {
    const driver = userEvent.setup();

    renderWithTranslations(<SupportChatWidget token="token" user={user} onUnauthorized={vi.fn()} />);

    await driver.click(screen.getByRole('button', { name: /open support chat/i }));
    const input = await screen.findByPlaceholderText(/write a message/i);

    await driver.type(input, 'Line one');
    await driver.keyboard('{Shift>}{Enter}{/Shift}Line two');

    expect(sendSupportMessage).not.toHaveBeenCalled();
    expect(input).toHaveValue('Line one\nLine two');
  });
});
