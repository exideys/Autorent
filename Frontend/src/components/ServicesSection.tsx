import { motion } from 'framer-motion';
import { Calendar, Car, MapPin, Shield, Star, Zap, type LucideIcon } from 'lucide-react';
import type { ServiceIconKey, ServiceItem } from '../types/site';

interface ServicesSectionProps {
  items: ServiceItem[];
}

const iconMap: Record<ServiceIconKey, LucideIcon> = {
  shield: Shield,
  zap: Zap,
  star: Star,
  car: Car,
  'map-pin': MapPin,
  calendar: Calendar,
};

const ServicesSection = ({ items }: ServicesSectionProps) => (
  <section id="services" className="py-16 px-4">
    <div className="max-w-7xl mx-auto">
      <motion.h2
        initial={{ opacity: 0, y: 50 }}
        whileInView={{ opacity: 1, y: 0 }}
        className="text-4xl font-bold text-center mb-12 text-cyan-400"
      >
        Premium Services
      </motion.h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
        {items.map((service, index) => {
          const Icon = iconMap[service.icon];
          return (
            <motion.div
              key={service.title}
              initial={{ opacity: 0, y: 50 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.08 }}
              whileHover={{ scale: 1.03 }}
              className="bg-white/10 backdrop-blur-lg rounded-3xl p-6 border border-cyan-500/20 hover:border-cyan-400/50 transition-all duration-300 shadow-lg hover:shadow-cyan-500/25"
            >
              <Icon size={48} className="text-cyan-400 mb-4" />
              <h3 className="text-xl font-semibold mb-2">{service.title}</h3>
              <p className="text-gray-300">{service.desc}</p>
            </motion.div>
          );
        })}
      </div>
    </div>
  </section>
);

export default ServicesSection;
