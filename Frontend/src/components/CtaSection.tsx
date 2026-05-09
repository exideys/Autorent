import { motion } from 'framer-motion';

const CtaSection = () => (
  <section className="py-16 px-4 bg-gradient-to-r from-cyan-900/20 to-violet-900/20">
    <div className="max-w-4xl mx-auto text-center">
      <motion.h2
        initial={{ opacity: 0, y: 50 }}
        whileInView={{ opacity: 1, y: 0 }}
        className="text-4xl font-bold mb-6 text-cyan-400"
      >
        Ready to Experience Luxury?
      </motion.h2>
      <motion.p
        initial={{ opacity: 0, y: 50 }}
        whileInView={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="text-xl text-gray-300 mb-8"
      >
        Join thousands of satisfied customers who trust AutoRent for their premium transportation needs.
      </motion.p>
      <motion.button
        initial={{ opacity: 0, y: 50 }}
        whileInView={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.4 }}
        type="button"
        className="bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 text-white font-semibold py-4 px-10 rounded-3xl text-lg transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-cyan-500/25"
      >
        Get Started Today
      </motion.button>
    </div>
  </section>
);

export default CtaSection;
