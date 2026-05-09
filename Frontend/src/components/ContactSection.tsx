import { motion } from 'framer-motion';
import type { ContactInfo } from '../types/site';

interface ContactSectionProps {
  contact: ContactInfo;
}

const ContactSection = ({ contact }: ContactSectionProps) => (
  <section className="py-16 px-4">
    <div className="max-w-4xl mx-auto">
      <motion.div
        initial={{ opacity: 0, y: 50 }}
        whileInView={{ opacity: 1, y: 0 }}
        className="bg-white/10 backdrop-blur-lg rounded-3xl p-8 border border-cyan-500/20 shadow-2xl"
      >
        <h2 className="text-3xl font-bold text-center mb-4 text-cyan-400">Contact AutoRent</h2>
        <p className="text-gray-300 text-center mb-8">
          Reach out to our team for booking help, VIP service requests, or any questions about our fleet.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-black/50 border border-cyan-500/20 rounded-3xl p-6">
            <h3 className="text-xl font-semibold text-cyan-300 mb-3">Email</h3>
            <a href={`mailto:${contact.email}`} className="text-gray-300 hover:text-cyan-300 transition-colors">
              {contact.email}
            </a>
          </div>
          <div className="bg-black/50 border border-cyan-500/20 rounded-3xl p-6">
            <h3 className="text-xl font-semibold text-cyan-300 mb-3">Phone</h3>
            <a href={`tel:${contact.phone.replace(/[^\d+]/g, '')}`} className="text-gray-300 hover:text-cyan-300 transition-colors">
              {contact.phone}
            </a>
          </div>
        </div>
      </motion.div>
    </div>
  </section>
);

export default ContactSection;
