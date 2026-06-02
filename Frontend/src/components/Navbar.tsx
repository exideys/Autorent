import { motion } from 'framer-motion';
import { Menu, X } from 'lucide-react';
import type { ReactNode } from 'react';
import type { PageKey, PageLink } from '../types/site';

interface NavbarProps {
  pages: PageLink[];
  activePage: PageKey;
  isMenuOpen: boolean;
  onToggleMenu: () => void;
  onNavigate: (page: PageKey) => void;
  actions?: ReactNode;
}

const Navbar = ({ pages, activePage, isMenuOpen, onToggleMenu, onNavigate, actions }: NavbarProps) => (
  <nav className="fixed top-0 w-full bg-black/80 backdrop-blur-md z-50 border-b border-cyan-500/20" aria-label="Primary">
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="grid h-16 grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 lg:gap-5">
        <motion.button
          type="button"
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          onClick={() => onNavigate('home')}
          className="whitespace-nowrap text-2xl font-bold text-cyan-400 transition-colors hover:text-cyan-300 focus:outline-none"
          aria-label="Go to home page"
        >
          AutoRent
        </motion.button>
        <div className="hidden min-w-0 items-center justify-center gap-4 xl:flex 2xl:gap-6">
          {pages.map((page) => (
            <button
              key={page.key}
              type="button"
              onClick={() => onNavigate(page.key)}
              aria-current={activePage === page.key ? 'page' : undefined}
              className={`whitespace-nowrap text-sm text-gray-300 transition-colors duration-300 2xl:text-base ${
                activePage === page.key ? 'text-cyan-400 font-semibold' : 'hover:text-cyan-400'
              }`}
            >
              {page.label}
            </button>
          ))}
        </div>
        <div className="flex min-w-0 items-center justify-end gap-2 sm:gap-3">
          {actions}
          <button
            type="button"
            className="text-cyan-400 xl:hidden"
            onClick={onToggleMenu}
            aria-label={isMenuOpen ? 'Close navigation menu' : 'Open navigation menu'}
            aria-expanded={isMenuOpen}
            aria-controls="mobile-navigation"
          >
            {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>
      </div>
      {isMenuOpen && (
        <motion.div
          id="mobile-navigation"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mt-2 rounded-lg bg-black/90 p-4 backdrop-blur-md xl:hidden"
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
