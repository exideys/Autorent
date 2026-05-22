import React from 'react';
import { motion } from 'framer-motion';
import { Car, MapPin, Calendar, Clock, Shield, Zap, Star, Phone, Mail, Menu, X, Gauge, Users } from 'lucide-react';

interface CarItem {
  id: number;
  name: string;
  category: string;
  location: string;
  seats: number;
  transmission: string;
  pricePerDay: number;
  image: string;
}

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/$/, '');

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

const App: React.FC = () => {
  const [isMenuOpen, setIsMenuOpen] = React.useState(false);
  const [activePage, setActivePage] = React.useState('home');
  const [cars, setCars] = React.useState<CarItem[]>([]);
  const [carsLoading, setCarsLoading] = React.useState(true);
  const [carsError, setCarsError] = React.useState<string | null>(null);

  const loadCars = React.useCallback(async () => {
    try {
      setCarsLoading(true);
      setCarsError(null);

      const response = await fetch(`${API_BASE_URL}/cars`);
      if (!response.ok) {
        throw new Error(`Backend returned ${response.status}`);
      }

      const data = await response.json();
      setCars(Array.isArray(data) ? data : []);
    } catch (error) {
      setCars([]);
      setCarsError(error instanceof Error ? error.message : 'Failed to load cars');
    } finally {
      setCarsLoading(false);
    }
  }, []);

  React.useEffect(() => {
    void loadCars();
  }, [loadCars]);

  const pages = [
    { key: 'home', label: 'Home' },
    { key: 'services', label: 'Services' },
    { key: 'showroom', label: 'Showroom' },
    { key: 'how-it-works', label: 'How It Works' },
    { key: 'why-choose-us', label: 'Why Choose Us' },
    { key: 'contact', label: 'Contact' },
  ];

  const pageConfig = {
    home: {
      title: 'Drive the Future',
      subtitle: 'Premium luxury vehicles at your fingertips. Experience unparalleled comfort and style.',
      buttonText: 'Explore Services',
      buttonAction: () => setActivePage('services'),
    },
    services: {
      title: 'Premium Services',
      subtitle: 'Discover the full range of services designed to make your rental experience seamless.',
      buttonText: 'Back to Overview',
      buttonAction: () => setActivePage('home'),
    },
    showroom: {
      title: 'Virtual Showroom',
      subtitle: 'Our showroom page is dedicated only to luxury vehicle previews and selection details.',
      buttonText: 'Back to Overview',
      buttonAction: () => setActivePage('home'),
    },
    'how-it-works': {
      title: 'How It Works',
      subtitle: 'A focused guide covering every step of your AutoRent booking experience.',
      buttonText: 'Back to Overview',
      buttonAction: () => setActivePage('home'),
    },
    'why-choose-us': {
      title: 'Why Choose AutoRent',
      subtitle: 'Everything that makes our luxury rental service stand apart.',
      buttonText: 'Back to Overview',
      buttonAction: () => setActivePage('home'),
    },
    contact: {
      title: 'Contact AutoRent',
      subtitle: 'Get in touch with our team for quick support and premium booking assistance.',
      buttonText: 'Back to Overview',
      buttonAction: () => setActivePage('home'),
    },
  };

  const activePageConfig = pageConfig[activePage as keyof typeof pageConfig];
  const isHome = activePage === 'home';

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-black to-gray-800 text-white overflow-x-hidden">
      <nav className="fixed top-0 w-full bg-black/80 backdrop-blur-md z-50 border-b border-cyan-500/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <motion.div
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              className="text-2xl font-bold text-cyan-400"
            >
              AutoRent
            </motion.div>
            <div className="hidden md:flex space-x-8">
              {pages.map((page) => (
                <button
                  key={page.key}
                  type="button"
                  onClick={() => setActivePage(page.key)}
                  className={`text-gray-300 transition-colors duration-300 ${activePage === page.key ? 'text-cyan-400 font-semibold' : 'hover:text-cyan-400'}`}
                >
                  {page.label}
                </button>
              ))}
            </div>
            <button
              type="button"
              className="md:hidden text-cyan-400"
              onClick={() => setIsMenuOpen(!isMenuOpen)}
            >
              {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
            </button>
          </div>
          {isMenuOpen && (
            <motion.div
              initial={{ opacity: 0, y: -20 }}
              animate={{ opacity: 1, y: 0 }}
              className="md:hidden bg-black/90 backdrop-blur-md rounded-lg mt-2 p-4"
            >
              {pages.map((page) => (
                <button
                  key={page.key}
                  type="button"
                  onClick={() => {
                    setActivePage(page.key);
                    setIsMenuOpen(false);
                  }}
                  className="block w-full text-left py-2 text-gray-300 hover:text-cyan-400 transition-colors duration-300"
                >
                  {page.label}
                </button>
              ))}
            </motion.div>
          )}
        </div>
      </nav>

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
              />
              <div className="absolute inset-0 bg-gradient-to-br from-black/80 via-black/55 to-black/80" />
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
            {activePageConfig.title}
          </motion.h1>
          <motion.p
            initial={{ opacity: 0, y: 50 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="text-xl md:text-2xl text-gray-300 mb-8"
          >
            {activePageConfig.subtitle}
          </motion.p>
          <motion.button
            initial={{ opacity: 0, y: 50 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 1, delay: 0.4 }}
            type="button"
            onClick={activePageConfig.buttonAction}
            className="bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 text-white font-semibold py-3 px-8 rounded-3xl transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-cyan-500/25"
          >
            {activePageConfig.buttonText}
          </motion.button>
        </div>
      </section>

      {isHome && (
        <section className="py-16 px-4">
          <div className="max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, y: 50 }}
              whileInView={{ opacity: 1, y: 0 }}
              className="bg-white/10 backdrop-blur-lg rounded-3xl p-8 border border-cyan-500/20 shadow-2xl"
            >
              <h2 className="text-3xl font-bold text-center mb-8 text-cyan-400">Book Your Ride</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <div className="space-y-2">
                  <label className="text-sm text-gray-300 flex items-center">
                    <MapPin size={16} className="mr-2 text-cyan-400" />
                    Pickup Location
                  </label>
                  <input
                    type="text"
                    placeholder="Enter location"
                    className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white placeholder-gray-500 focus:border-cyan-400 focus:outline-none transition-colors"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm text-gray-300 flex items-center">
                    <Calendar size={16} className="mr-2 text-cyan-400" />
                    Pickup Date
                  </label>
                  <input
                    type="date"
                    className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm text-gray-300 flex items-center">
                    <Calendar size={16} className="mr-2 text-cyan-400" />
                    Return Date
                  </label>
                  <input
                    type="date"
                    className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm text-gray-300 flex items-center">
                    <Clock size={16} className="mr-2 text-cyan-400" />
                    Pickup Time
                  </label>
                  <input
                    type="time"
                    className="w-full bg-black/50 border border-cyan-500/30 rounded-2xl px-4 py-3 text-white focus:border-cyan-400 focus:outline-none transition-colors"
                  />
                </div>
              </div>
              <div className="text-center mt-8">
                <button className="bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 text-white font-semibold py-3 px-8 rounded-3xl transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-cyan-500/25">
                  Search Vehicles
                </button>
              </div>
            </motion.div>
          </div>
        </section>
      )}

      {(isHome || activePage === 'services') && (
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
            {[
              { icon: Shield, title: 'VIP Concierge', desc: 'Personalized service from booking to return' },
              { icon: Zap, title: 'Instant Booking', desc: 'Reserve your vehicle in seconds' },
              { icon: Star, title: 'Luxury Fleet', desc: 'Access to premium and exotic cars' },
              { icon: Car, title: '24/7 Support', desc: 'Round-the-clock assistance' },
              { icon: MapPin, title: 'Global Network', desc: 'Pickup and drop-off worldwide' },
              { icon: Calendar, title: 'Flexible Terms', desc: 'Custom rental periods and pricing' },
            ].map((service, index) => (
              <motion.div
                key={service.title}
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                whileHover={{ scale: 1.05 }}
                className="bg-white/10 backdrop-blur-lg rounded-3xl p-6 border border-cyan-500/20 hover:border-cyan-400/50 transition-all duration-300 shadow-lg hover:shadow-cyan-500/25"
              >
                <service.icon size={48} className="text-cyan-400 mb-4" />
                <h3 className="text-xl font-semibold mb-2">{service.title}</h3>
                <p className="text-gray-300">{service.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>
      )}

      {(isHome || activePage === 'showroom') && (
        <section id="showroom" className="py-16 px-4">
        <div className="max-w-7xl mx-auto">
          <motion.h2
            initial={{ opacity: 0, y: 50 }}
            whileInView={{ opacity: 1, y: 0 }}
            className="text-4xl font-bold text-center mb-8 text-cyan-400"
          >
            Virtual Showroom
          </motion.h2>
          <div className="bg-white/5 backdrop-blur-xl border border-cyan-500/20 rounded-3xl p-8 shadow-2xl">
            <div className="mb-8 flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
              <p aria-live="polite" className="text-gray-300 text-lg">
                {carsLoading && 'Loading vehicles from TiDB Cloud...'}
                {!carsLoading && carsError && `Could not load vehicles: ${carsError}`}
                {!carsLoading && !carsError && `${cars.length} vehicles loaded from TiDB Cloud.`}
              </p>
              <button
                type="button"
                onClick={loadCars}
                className="self-start rounded-2xl border border-cyan-400/40 px-4 py-2 text-sm text-cyan-200 transition-colors hover:bg-cyan-400/10 md:self-auto"
              >
                Refresh
              </button>
            </div>

            {carsLoading ? (
              <div className="rounded-3xl border border-cyan-500/10 bg-black/40 py-12 text-center text-gray-300">
                Fetching live inventory...
              </div>
            ) : carsError ? (
              <div className="rounded-3xl border border-red-500/30 bg-red-950/20 p-6 text-red-100">
                Backend or TiDB is not responding. Check `docker compose logs -f backend` and `/api/db/health`.
              </div>
            ) : cars.length === 0 ? (
              <div className="rounded-3xl border border-cyan-500/10 bg-black/40 py-12 text-center">
                <p className="text-xl text-gray-300">No vehicles found.</p>
                <p className="mt-2 text-sm text-gray-400">The backend is connected, but the cars table is empty.</p>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {cars.map((car) => (
                  <motion.article
                    key={car.id}
                    whileHover={{ scale: 1.02 }}
                    className="overflow-hidden rounded-3xl border border-cyan-500/20 bg-black/45 shadow-lg transition-all duration-300 hover:shadow-cyan-500/25"
                  >
                    <div className="relative h-44">
                      <img
                        src={car.image || '/hero-main.png'}
                        alt={car.name}
                        className="h-full w-full object-cover"
                        onError={(event) => {
                          event.currentTarget.onerror = null;
                          event.currentTarget.src = '/hero-main.png';
                        }}
                      />
                      <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
                      <span className="absolute left-3 top-3 rounded-full border border-cyan-300/30 bg-cyan-500/20 px-3 py-1 text-xs text-cyan-100">
                        {car.category}
                      </span>
                    </div>
                    <div className="p-5">
                      <h3 className="mb-2 text-xl font-semibold">{car.name}</h3>
                      <div className="mb-4 space-y-2 text-sm text-gray-300">
                        <div className="flex items-center gap-2">
                          <MapPin size={15} className="text-cyan-300" />
                          <span>{car.location}</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Users size={15} className="text-cyan-300" />
                          <span>{car.seats} seats</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Gauge size={15} className="text-cyan-300" />
                          <span>{car.transmission}</span>
                        </div>
                      </div>
                      <div className="flex items-center justify-between">
                        <p className="font-semibold text-cyan-300">
                          {currencyFormatter.format(car.pricePerDay)} / day
                        </p>
                        <button
                          type="button"
                          className="rounded-2xl bg-gradient-to-r from-cyan-500 to-violet-600 px-4 py-2 text-sm transition-colors hover:from-cyan-600 hover:to-violet-700"
                        >
                          Book Now
                        </button>
                      </div>
                    </div>
                  </motion.article>
                ))}
              </div>
            )}
          </div>
        </div>
      </section>
      )}

      {(isHome || activePage === 'how-it-works') && (
        <section id="how-it-works" className="py-16 px-4 bg-gradient-to-r from-cyan-900/10 to-violet-900/10">
        <div className="max-w-7xl mx-auto">
          <motion.h2
            initial={{ opacity: 0, y: 50 }}
            whileInView={{ opacity: 1, y: 0 }}
            className="text-4xl font-bold text-center mb-12 text-cyan-400"
          >
            How It Works
          </motion.h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              { step: '01', title: 'Choose & Book', desc: 'Select your preferred vehicle and dates' },
              { step: '02', title: 'Pickup', desc: 'Meet at your chosen location' },
              { step: '03', title: 'Enjoy', desc: 'Drive with confidence and style' },
            ].map((item, index) => (
              <motion.div
                key={item.step}
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.2 }}
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
      )}

      {(isHome || activePage === 'why-choose-us') && (
      <section id="why-choose-us" className="py-16 px-4">
        <div className="max-w-7xl mx-auto">
          <motion.h2
            initial={{ opacity: 0, y: 50 }}
            whileInView={{ opacity: 1, y: 0 }}
            className="text-4xl font-bold text-center mb-12 text-cyan-400"
          >
            Why Choose AutoRent
          </motion.h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {[
              { title: 'Unmatched Quality', desc: 'Every vehicle in our fleet undergoes rigorous maintenance and inspection.' },
              { title: 'Competitive Pricing', desc: 'Transparent pricing with no hidden fees or surprises.' },
              { title: 'Customer Satisfaction', desc: 'Rated 5-star by thousands of satisfied customers worldwide.' },
              { title: 'Innovation', desc: 'Cutting-edge technology for seamless booking and management.' },
            ].map((item, index) => (
              <motion.div
                key={item.title}
                initial={{ opacity: 0, x: index % 2 === 0 ? -50 : 50 }}
                whileInView={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
                className="bg-white/10 backdrop-blur-lg rounded-3xl p-6 border border-cyan-500/20"
              >
                <h3 className="text-2xl font-semibold mb-2 text-cyan-400">{item.title}</h3>
                <p className="text-gray-300">{item.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      )}

      {activePage === 'contact' && (
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
                  <p className="text-gray-300">info@autorent.com</p>
                </div>
                <div className="bg-black/50 border border-cyan-500/20 rounded-3xl p-6">
                  <h3 className="text-xl font-semibold text-cyan-300 mb-3">Phone</h3>
                  <p className="text-gray-300">+1 (555) 123-4567</p>
                </div>
              </div>
            </motion.div>
          </div>
        </section>
      )}

      {isHome && (
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
            className="bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 text-white font-semibold py-4 px-10 rounded-3xl text-lg transition-all duration-300 transform hover:scale-105 shadow-lg hover:shadow-cyan-500/25"
          >
            Get Started Today
          </motion.button>
        </div>
      </section>
      )}

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
                <div className="flex items-center">
                  <Phone size={16} className="mr-2 text-cyan-400" />
                  +1 (555) 123-4567
                </div>
                <div className="flex items-center">
                  <Mail size={16} className="mr-2 text-cyan-400" />
                  info@autorent.com
                </div>
              </div>
            </div>
          </div>
          <div className="border-t border-cyan-500/20 mt-8 pt-8 text-center text-gray-400">
            <p>&copy; 2024 AutoRent. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default App;
