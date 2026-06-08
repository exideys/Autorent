import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { TranslationProvider } from '../i18n/TranslationContext';

export const renderWithTranslations = (ui: ReactElement) => render(<TranslationProvider>{ui}</TranslationProvider>);
