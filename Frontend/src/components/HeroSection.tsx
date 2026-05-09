import { motion } from 'framer-motion';
import type { PageContent } from '../types/site';

interface HeroSectionProps {
  content: PageContent;
  buttonText: string;
  onButtonClick: () => void;
  isHome: boolean;
}

const HeroSection = ({ content, buttonText, onButtonClick, isHome }: HeroSectionProps) => (
  <section className="relative min-h-screen flex items-center justify-center pt-16">
    <div className="absolute inset-0 overflow-hidden">
      {isHome ? (
        <>
          <motion.img
            src="/hero-main.png"
            alt="AutoRent luxury showroom"
            initial={{ scale: 1 }}
            animate={{ scale: 1.12 }}
            transition={{
              duration: 22,
              ease: 'easeInOut',
              repeat: Infinity,
              repeatType: 'reverse',
            }}
            className="absolute inset-0 h-full w-full object-cover"
            onError={(event) => {
              event.currentTarget.style.display = 'none';
            }}
          />
          <div className="absolute inset-0 bg-gradient-to-br from-black/80 via-black/55 to-black/80" />
          <motion.div
            animate={{
              backgroundPosition: ['0% 0%', '100% 100%'],
            }}
            transition={{
              duration: 18,
              repeat: Infinity,
              repeatType: 'reverse',
            }}
            className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(34,211,238,0.18),transparent_55%),radial-gradient(circle_at_80%_10%,rgba(168,85,247,0.16),transparent_50%)]"
          />
        </>
      ) : (
        <>
          <motion.div
            animate={{
              backgroundPosition: ['0% 0%', '100% 100%'],
            }}
            transition={{
              duration: 20,
              repeat: Infinity,
              repeatType: 'reverse',
            }}
            className="absolute inset-0 bg-gradient-to-br from-cyan-900/20 via-violet-900/20 to-blue-900/20"
          />
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(0,255,255,0.1),transparent_50%)]" />
        </>
      )}
      <div className="absolute bottom-0 left-0 right-0 h-32 bg-gradient-to-t from-black to-transparent" />
    </div>
    <div className="relative z-10 text-center max-w-4xl mx-auto px-4">
      <motion.h1
        initial={{ opacity: 0, y: 50 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 1 }}
        className="text-5xl md:text-7xl font-bold mb-6 bg-gradient-to-r from-cyan-400 via-blue-500 to-violet-600 bg-clip-text text-transparent"
      >
        {content.title}
      </motion.h1>
      <motion.p
        initial={{ opacity: 0, y: 50 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 1, delay: 0.2 }}
        className="text-xl md:text-2xl text-gray-300 mb-8"
      >
        {content.subtitle}
      </motion.p>
      <motion.button
        initial={{ opacity: 0, y: 50 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 1, delay: 0.4 }}
        type="button"
        onClick={onButtonClick}
        className="bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 text-white font-semibold py-3 px-8 rounded-3xl transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-cyan-500/25"
      >
        {buttonText}
      </motion.button>
    </div>
  </section>
);

export default HeroSection;
