import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';
import { translateTexts } from '../lib/api';

export type Language = 'ENG' | 'UA';

type TranslationText = string | null | undefined;

interface TranslationContextValue {
  language: Language;
  setLanguage: (language: Language) => void;
  t: (text: TranslationText) => string;
  requestTranslations: (texts: readonly TranslationText[]) => void;
  isTranslating: boolean;
  error: string;
  clearError: () => void;
}

interface TranslationProviderProps {
  children: ReactNode;
}

const batchSize = 50;
const targetLang = 'UK';
const unavailableMessage = 'Translation service is temporarily unavailable. English text is still shown.';
const TranslationContext = createContext<TranslationContextValue | null>(null);

const normalizeText = (text: TranslationText) => {
  if (text === null || text === undefined) {
    return '';
  }

  return String(text);
};

const uniqueTexts = (texts: readonly TranslationText[]) => {
  const seen = new Set<string>();
  const values: string[] = [];

  texts.forEach((item) => {
    const text = normalizeText(item);
    if (text.trim() === '' || seen.has(text)) {
      return;
    }
    seen.add(text);
    values.push(text);
  });

  return values;
};

export const TranslationProvider = ({ children }: TranslationProviderProps) => {
  const [language, setLanguageState] = useState<Language>('ENG');
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const [queue, setQueue] = useState<string[]>([]);
  const [isTranslating, setIsTranslating] = useState(false);
  const [error, setError] = useState('');
  const pendingRef = useRef(new Set<string>());
  const failedRef = useRef(new Set<string>());
  const inFlightRef = useRef(false);
  const mountedRef = useRef(true);
  const serviceUnavailableRef = useRef(false);

  useEffect(
    () => () => {
      mountedRef.current = false;
    },
    [],
  );

  const setLanguage = useCallback((nextLanguage: Language) => {
    setLanguageState(nextLanguage);
    setError('');
    failedRef.current.clear();
    serviceUnavailableRef.current = false;

    if (nextLanguage === 'ENG') {
      pendingRef.current.clear();
      setQueue([]);
      setIsTranslating(false);
    }
  }, []);

  const requestTranslations = useCallback(
    (texts: readonly TranslationText[]) => {
      if (language !== 'UA' || serviceUnavailableRef.current) {
        return;
      }

      const nextTexts = uniqueTexts(texts).filter(
        (text) => !translations[text] && !pendingRef.current.has(text) && !failedRef.current.has(text),
      );
      if (nextTexts.length === 0) {
        return;
      }

      nextTexts.forEach((text) => pendingRef.current.add(text));
      setQueue((current) => {
        const queuedTexts = new Set(current);
        const merged = [...current];

        nextTexts.forEach((text) => {
          if (!queuedTexts.has(text)) {
            merged.push(text);
          }
        });

        return merged;
      });
    },
    [language, translations],
  );

  const flushQueue = useCallback(() => {
    if (language !== 'UA' || queue.length === 0 || inFlightRef.current) {
      return;
    }

    const batch = queue.slice(0, batchSize);
    const batchSet = new Set(batch);
    inFlightRef.current = true;
    setIsTranslating(true);
    setQueue((current) => current.filter((text) => !batchSet.has(text)));

    translateTexts(batch, targetLang)
      .then((response) => {
        if (!mountedRef.current || response.translations.length !== batch.length) {
          throw new Error('Invalid translation response.');
        }

        setTranslations((current) => {
          const nextTranslations = { ...current };
          batch.forEach((text, index) => {
            nextTranslations[text] = response.translations[index];
          });
          return nextTranslations;
        });
        setError('');
      })
      .catch(() => {
        if (mountedRef.current) {
          serviceUnavailableRef.current = true;
          queue.forEach((text) => failedRef.current.add(text));
          pendingRef.current.clear();
          setQueue([]);
          setError(unavailableMessage);
        }
      })
      .finally(() => {
        batch.forEach((text) => pendingRef.current.delete(text));
        inFlightRef.current = false;
        if (mountedRef.current) {
          setIsTranslating(false);
        }
      });
  }, [language, queue]);

  useEffect(() => {
    flushQueue();
  }, [flushQueue, isTranslating]);

  const t = useCallback(
    (text: TranslationText) => {
      const normalizedText = normalizeText(text);
      if (language !== 'UA' || normalizedText.trim() === '') {
        return normalizedText;
      }

      return translations[normalizedText] || normalizedText;
    },
    [language, translations],
  );

  const clearError = useCallback(() => setError(''), []);

  const value = useMemo(
    () => ({
      language,
      setLanguage,
      t,
      requestTranslations,
      isTranslating,
      error,
      clearError,
    }),
    [clearError, error, isTranslating, language, requestTranslations, setLanguage, t],
  );

  return <TranslationContext.Provider value={value}>{children}</TranslationContext.Provider>;
};

export const useTranslation = (texts: readonly TranslationText[] = []) => {
  const context = useContext(TranslationContext);
  if (!context) {
    throw new Error('useTranslation must be used within TranslationProvider');
  }

  const textKey = uniqueTexts(texts).join('\u001f');
  const normalizedTexts = useMemo(() => uniqueTexts(texts), [textKey]);

  useEffect(() => {
    context.requestTranslations(normalizedTexts);
  }, [context, normalizedTexts]);

  return context;
};
