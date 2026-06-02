import { motion } from 'framer-motion';
import { useTranslation } from '../i18n/TranslationContext';
import type { BenefitItem } from '../types/site';

interface WhyChooseUsSectionProps {
  items: BenefitItem[];
}

const WhyChooseUsSection = ({ items }: WhyChooseUsSectionProps) => {
  const { t } = useTranslation(['Why Choose AutoRent']);

  return (
    <section id="why-choose-us" className="py-16 px-4">
      <div className="max-w-7xl mx-auto">
        <motion.h2
          initial={{ opacity: 0, y: 50 }}
          whileInView={{ opacity: 1, y: 0 }}
          className="text-4xl font-bold text-center mb-12 text-cyan-400"
        >
          {t('Why Choose AutoRent')}
        </motion.h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
          {items.map((item, index) => (
            <motion.div
              key={item.title}
              initial={{ opacity: 0, x: index % 2 === 0 ? -40 : 40 }}
              whileInView={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.08 }}
              className="bg-white/10 backdrop-blur-lg rounded-3xl p-6 border border-cyan-500/20"
            >
              <h3 className="text-2xl font-semibold mb-2 text-cyan-400">{item.title}</h3>
              <p className="text-gray-300">{item.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default WhyChooseUsSection;
