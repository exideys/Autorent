import { motion } from 'framer-motion';
import { Menu, X } from 'lucide-react';
import type { PageKey, PageLink } from '../types/site';

interface NavbarProps {
  pages: PageLink[];
  activePage: PageKey;
  isMenuOpen: boolean;
  onToggleMenu: () => void;
  onNavigate: (page: PageKey) => void;
}

const Navbar = ({ pages, activePage, isMenuOpen, onToggleMenu, onNavigate }: NavbarProps) => (
  <nav className="fixed top-0 w-full bg-black/80 backdrop-blur-md z-50 border-b border-cyan-500/20" aria-label="Primary">
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
              onClick={() => onNavigate(page.key)}
              aria-current={activePage === page.key ? 'page' : undefined}
              className={`text-gray-300 transition-colors duration-300 ${activePage === page.key ? 'text-cyan-400 font-semibold' : 'hover:text-cyan-400'}`}
            >
              {page.label}
            </button>
          ))}
        </div>
        <button
          type="button"
          className="md:hidden text-cyan-400"
          onClick={onToggleMenu}
          aria-label={isMenuOpen ? 'Close navigation menu' : 'Open navigation menu'}
          aria-expanded={isMenuOpen}
          aria-controls="mobile-navigation"
        >
          {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
      </div>
      {isMenuOpen && (
        <motion.div
          id="mobile-navigation"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="md:hidden bg-black/90 backdrop-blur-md rounded-lg mt-2 p-4"
        >
          {pages.map((page) => (
            <button
              key={page.key}
              type="button"
              onClick={() => onNavigate(page.key)}
              aria-current={activePage === page.key ? 'page' : undefined}
              className="block w-full text-left py-2 text-gray-300 hover:text-cyan-400 transition-colors duration-300"
            >
              {page.label}
            </button>
          ))}
        </motion.div>
      )}
    </div>
  </nav>
);

export default Navbar;
