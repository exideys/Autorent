import { Languages } from 'lucide-react';
import { useTranslation, type Language } from '../i18n/TranslationContext';

const languages: Language[] = ['ENG', 'UA'];

const LanguageToggle = () => {
  const { language, setLanguage, isTranslating } = useTranslation();

  return (
    <div className="relative inline-flex shrink-0">
      <div
        className="inline-flex h-10 items-center rounded-lg border border-cyan-500/30 bg-black/50 p-1 text-sm font-semibold text-cyan-100"
        aria-label="Language"
      >
        <Languages size={16} className="mx-2 hidden text-cyan-300 sm:block" />
        {languages.map((item, index) => (
          <div key={item} className="flex items-center">
            <button
              type="button"
              onClick={() => setLanguage(item)}
              className={`rounded-md px-2 py-1 transition-colors ${
                language === item ? 'bg-cyan-500 text-black' : 'text-gray-300 hover:bg-white/10 hover:text-cyan-100'
              }`}
              aria-pressed={language === item}
            >
              {item}
            </button>
            {index === 0 && <span className="px-1 text-gray-500" aria-hidden="true">/</span>}
          </div>
        ))}
      </div>
      <span
        className={`pointer-events-none absolute -right-1 top-1/2 h-2 w-2 -translate-y-1/2 rounded-full bg-cyan-300 transition-opacity ${
          language === 'UA' && isTranslating ? 'opacity-100 motion-safe:animate-pulse' : 'opacity-0'
        }`}
        title={language === 'UA' && isTranslating ? 'Translation loading' : undefined}
        aria-hidden="true"
      />
    </div>
  );
};

export default LanguageToggle;
