import { motion } from 'framer-motion';
import { useTranslation } from '../i18n/TranslationContext';
import type { HowItWorksStep } from '../types/site';

interface HowItWorksSectionProps {
  items: HowItWorksStep[];
}

const HowItWorksSection = ({ items }: HowItWorksSectionProps) => {
  const { t } = useTranslation(['How It Works']);

  return (
    <section id="how-it-works" className="py-16 px-4 bg-gradient-to-r from-cyan-900/10 to-violet-900/10">
      <div className="max-w-7xl mx-auto">
        <motion.h2
          initial={{ opacity: 0, y: 50 }}
          whileInView={{ opacity: 1, y: 0 }}
          className="text-4xl font-bold text-center mb-12 text-cyan-400"
        >
          {t('How It Works')}
        </motion.h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {items.map((item, index) => (
            <motion.div
              key={item.step}
              initial={{ opacity: 0, y: 50 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.15 }}
              className="text-center"
            >
              <div className="bg-gradient-to-r from-cyan-500 to-violet-600 rounded-full w-16 h-16 flex items-center justify-center text-2xl font-bold mx-auto mb-4">
                {item.step}
              </div>
              <h3 className="text-2xl font-semibold mb-2">{item.title}</h3>
              <p className="text-gray-300">{item.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default HowItWorksSection;
