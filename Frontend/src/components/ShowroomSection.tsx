import { motion } from 'framer-motion';
import { Gauge, MapPin, Users } from 'lucide-react';
import type { CarItem, ShowroomHint } from '../types/site';

interface ShowroomSectionProps {
  cars: CarItem[];
  hints: ShowroomHint[];
  resultMessage: string;
}

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
});

const ShowroomSection = ({ cars, hints, resultMessage }: ShowroomSectionProps) => (
  <section id="showroom" className="py-16 px-4">
    <div className="max-w-7xl mx-auto">
      <motion.h2
        initial={{ opacity: 0, y: 50 }}
        whileInView={{ opacity: 1, y: 0 }}
        className="text-4xl font-bold text-center mb-4 text-cyan-400"
      >
        Virtual Showroom
      </motion.h2>
      <p className="text-center text-gray-300 mb-8 max-w-3xl mx-auto">
        Choose your vehicle by class, location, and available dates. Everything below is ready for quick booking.
      </p>
      <div className="bg-white/5 backdrop-blur-xl border border-cyan-500/20 rounded-3xl p-8 shadow-2xl">
        <p aria-live="polite" className="text-cyan-200 mb-6">
          {resultMessage}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          {hints.map((item) => (
            <div key={item.title} className="bg-black/50 border border-cyan-500/10 rounded-3xl p-5">
              <h3 className="text-lg font-semibold text-cyan-300 mb-2">{item.title}</h3>
              <p className="text-gray-400">{item.desc}</p>
            </div>
          ))}
        </div>

        {cars.length === 0 ? (
          <div className="text-center py-10 bg-black/40 border border-cyan-500/10 rounded-3xl">
            <p className="text-xl text-gray-300">No vehicles loaded yet.</p>
            <p className="text-sm text-gray-400 mt-2">Showroom inventory will be fetched from the database.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {cars.map((car) => (
              <motion.article
                key={car.id}
                whileHover={{ scale: 1.02 }}
                className="bg-black/45 border border-cyan-500/20 rounded-3xl overflow-hidden shadow-lg hover:shadow-cyan-500/25 transition-all duration-300"
              >
                <div className="relative h-44">
                  <img
                    src={car.image}
                    alt={car.name}
                    className="h-full w-full object-cover"
                    onError={(event) => {
                      event.currentTarget.onerror = null;
                      event.currentTarget.src = '/hero-main.png';
                    }}
                  />
                  <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
                  <span className="absolute top-3 left-3 text-xs px-3 py-1 rounded-full bg-cyan-500/20 border border-cyan-300/30 text-cyan-100">
                    {car.category}
                  </span>
                </div>
                <div className="p-5">
                  <h3 className="text-xl font-semibold mb-2">{car.name}</h3>
                  <div className="space-y-2 text-sm text-gray-300 mb-4">
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
                    <p className="text-cyan-300 font-semibold">{currencyFormatter.format(car.pricePerDay)} / day</p>
                    <button
                      type="button"
                      className="text-sm px-4 py-2 rounded-2xl bg-gradient-to-r from-cyan-500 to-violet-600 hover:from-cyan-600 hover:to-violet-700 transition-colors"
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
);

export default ShowroomSection;
