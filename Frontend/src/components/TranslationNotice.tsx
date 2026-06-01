import { X } from 'lucide-react';
import { useTranslation } from '../i18n/TranslationContext';

const TranslationNotice = () => {
  const { language, error, clearError } = useTranslation();

  if (language !== 'UA' || !error) {
    return null;
  }

  return (
    <div
      className="fixed right-4 top-20 z-[90] max-w-sm rounded-lg border border-amber-300/30 bg-amber-500/15 px-4 py-3 text-sm text-amber-50 shadow-2xl shadow-black/30 backdrop-blur"
      role="status"
    >
      <div className="flex items-start gap-3">
        <p>{error}</p>
        <button
          type="button"
          onClick={clearError}
          className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-amber-100 hover:bg-white/10"
          aria-label="Dismiss translation message"
        >
          <X size={15} />
        </button>
      </div>
    </div>
  );
};

export default TranslationNotice;
