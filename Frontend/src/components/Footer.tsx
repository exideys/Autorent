import { Mail, Phone } from 'lucide-react';
import type { ContactInfo } from '../types/site';

interface FooterProps {
  contact: ContactInfo;
}

const Footer = ({ contact }: FooterProps) => (
  <footer className="py-12 px-4 border-t border-cyan-500/20">
    <div className="max-w-7xl mx-auto">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
        <div>
          <h3 className="text-2xl font-bold text-cyan-400 mb-4">AutoRent</h3>
          <p className="text-gray-300">Driving the future of luxury transportation.</p>
        </div>
        <div>
          <h4 className="text-lg font-semibold mb-4">Services</h4>
          <ul className="space-y-2 text-gray-300">
            <li>Luxury Cars</li>
            <li>VIP Concierge</li>
            <li>Global Network</li>
            <li>24/7 Support</li>
          </ul>
        </div>
        <div>
          <h4 className="text-lg font-semibold mb-4">Company</h4>
          <ul className="space-y-2 text-gray-300">
            <li>About Us</li>
            <li>Careers</li>
            <li>Press</li>
            <li>Contact</li>
          </ul>
        </div>
        <div>
          <h4 className="text-lg font-semibold mb-4">Contact</h4>
          <div className="space-y-2 text-gray-300">
            <a href={`tel:${contact.phone.replace(/[^\d+]/g, '')}`} className="flex items-center hover:text-cyan-300 transition-colors">
              <Phone size={16} className="mr-2 text-cyan-400" />
              {contact.phone}
            </a>
            <a href={`mailto:${contact.email}`} className="flex items-center hover:text-cyan-300 transition-colors">
              <Mail size={16} className="mr-2 text-cyan-400" />
              {contact.email}
            </a>
          </div>
        </div>
      </div>
      <div className="border-t border-cyan-500/20 mt-8 pt-8 text-center text-gray-400">
        <p>&copy; {new Date().getFullYear()} AutoRent. All rights reserved.</p>
      </div>
    </div>
  </footer>
);

export default Footer;
