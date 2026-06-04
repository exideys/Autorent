import { BriefcaseBusiness, CarFront, CheckCircle2, MapPin, Sparkles } from 'lucide-react';
import { motion } from 'framer-motion';
import { useTranslation } from '../i18n/TranslationContext';

const principles = [
  {
    icon: CarFront,
    title: 'Luxury and business fleet',
    desc: 'Choose from premium sedans, SUVs, sport models, and executive vehicles prepared for city drives, meetings, events, and weekends.',
  },
  {
    icon: Sparkles,
    title: 'Concierge-style support',
    desc: 'Our team helps coordinate airport, hotel, office, and event pickups so every handoff feels simple, punctual, and polished.',
  },
  {
    icon: CheckCircle2,
    title: 'Transparent rental process',
    desc: 'Clear pricing, deposits, availability, and booking details keep every rental easy to understand before the keys are handed over.',
  },
];

const highlights = [
  'Luxury and business fleet',
  'Airport, hotel, event, and city travel',
  'Concierge booking assistance',
  'Clear terms from request to return',
];

const AboutUsSection = () => {
  const { t } = useTranslation([
    'About AutoRent',
    'Premium mobility for every important arrival',
    'AutoRent brings together a carefully selected luxury and business fleet with attentive service for clients who value comfort, timing, and confidence on the road.',
    'From airport arrivals and hotel transfers to executive meetings, events, and city travel, we make premium car rental feel clear, fast, and personal.',
    'What guides our service',
    'Built around your trip',
    'Whether you need a refined business sedan, a spacious SUV, or a statement car for a special route, AutoRent matches the vehicle and service flow to your schedule.',
    'Luxury and business fleet',
    'Airport, hotel, event, and city travel',
    'Concierge booking assistance',
    'Clear terms from request to return',
    ...principles.flatMap((item) => [item.title, item.desc]),
  ]);

  return (
    <section className="px-4 py-16">
      <div className="mx-auto max-w-7xl">
        <div className="grid gap-10 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
          <motion.div initial={{ opacity: 0, y: 24 }} whileInView={{ opacity: 1, y: 0 }} className="max-w-3xl">
            <p className="inline-flex items-center gap-2 text-sm font-semibold uppercase tracking-[0.24em] text-cyan-300">
              <BriefcaseBusiness size={17} />
              {t('About AutoRent')}
            </p>
            <h2 className="mt-3 text-4xl font-bold text-white md:text-5xl">{t('Premium mobility for every important arrival')}</h2>
            <p className="mt-5 text-lg leading-8 text-gray-300">
              {t(
                'AutoRent brings together a carefully selected luxury and business fleet with attentive service for clients who value comfort, timing, and confidence on the road.',
              )}
            </p>
            <p className="mt-4 leading-7 text-gray-400">
              {t(
                'From airport arrivals and hotel transfers to executive meetings, events, and city travel, we make premium car rental feel clear, fast, and personal.',
              )}
            </p>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 24 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.08 }}
            className="rounded-xl border border-cyan-500/20 bg-black/40 p-6 shadow-2xl shadow-black/25"
          >
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Built around your trip')}</p>
            <p className="mt-4 text-gray-300">
              {t(
                'Whether you need a refined business sedan, a spacious SUV, or a statement car for a special route, AutoRent matches the vehicle and service flow to your schedule.',
              )}
            </p>
            <div className="mt-6 grid gap-3 sm:grid-cols-2">
              {highlights.map((item) => (
                <div key={item} className="flex items-start gap-3 rounded-lg border border-cyan-500/10 bg-white/5 p-4 text-sm text-gray-200">
                  <MapPin size={16} className="mt-0.5 shrink-0 text-cyan-300" />
                  <span>{t(item)}</span>
                </div>
              ))}
            </div>
          </motion.div>
        </div>

        <div className="mt-12">
          <h3 className="text-2xl font-semibold text-white">{t('What guides our service')}</h3>
          <div className="mt-6 grid gap-5 md:grid-cols-3">
            {principles.map(({ icon: Icon, title, desc }, index) => (
              <motion.article
                key={title}
                initial={{ opacity: 0, y: 22 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.06 }}
                className="rounded-xl border border-cyan-500/20 bg-black/45 p-6 shadow-lg shadow-black/20"
              >
                <Icon size={34} className="text-cyan-300" />
                <h4 className="mt-5 text-xl font-semibold text-white">{t(title)}</h4>
                <p className="mt-3 text-sm leading-6 text-gray-300">{t(desc)}</p>
              </motion.article>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default AboutUsSection;
