import { useCallback, useEffect, useMemo, useState } from 'react';
import AdminDashboard from './components/AdminDashboard';
import AICarAssistant from './components/AICarAssistant';
import AuthMenu from './components/AuthMenu';
import ContactSection from './components/ContactSection';
import CtaSection from './components/CtaSection';
import Footer from './components/Footer';
import HeroSection from './components/HeroSection';
import HowItWorksSection from './components/HowItWorksSection';
import Navbar from './components/Navbar';
import ServicesSection from './components/ServicesSection';
import ShowroomSection from './components/ShowroomSection';
import WhyChooseUsSection from './components/WhyChooseUsSection';
import {
  benefits,
  contactInfo,
  howItWorksSteps,
  pageContent,
  pages,
  services,
} from './data/siteData';
import { getCurrentUser, listPublicCars } from './lib/api';
import type { AuthResponse, Car } from './types/api';
import type { PageKey } from './types/site';

const authStorageKey = 'autorent.auth';

const canUseStorage = () => typeof window !== 'undefined' && typeof window.localStorage !== 'undefined';

const readStoredAuth = (): AuthResponse | null => {
  if (!canUseStorage()) {
    return null;
  }

  const rawAuth = window.localStorage.getItem(authStorageKey);
  if (!rawAuth) {
    return null;
  }

  try {
    return JSON.parse(rawAuth) as AuthResponse;
  } catch {
    window.localStorage.removeItem(authStorageKey);
    return null;
  }
};

const saveAuth = (auth: AuthResponse) => {
  if (canUseStorage()) {
    window.localStorage.setItem(authStorageKey, JSON.stringify(auth));
  }
};

const clearStoredAuth = () => {
  if (canUseStorage()) {
    window.localStorage.removeItem(authStorageKey);
  }
};

const App = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [activePage, setActivePage] = useState<PageKey>('home');
  const [cars, setCars] = useState<Car[]>([]);
  const [isCarsLoading, setIsCarsLoading] = useState(true);
  const [carsError, setCarsError] = useState('');
  const [auth, setAuth] = useState<AuthResponse | null>(() => readStoredAuth());
  const [isSessionLoading, setIsSessionLoading] = useState(false);

  const user = auth?.user ?? null;
  const isAdminPage = activePage === 'admin';
  const isHome = activePage === 'home';

  const navPages = useMemo(
    () => (user?.role === 'admin' ? [...pages, { key: 'admin' as PageKey, label: 'Admin' }] : pages),
    [user?.role],
  );

  const loadCars = useCallback(async () => {
    setIsCarsLoading(true);
    setCarsError('');

    try {
      const loadedCars = await listPublicCars();
      setCars(loadedCars);
    } catch (loadError) {
      setCarsError(loadError instanceof Error ? loadError.message : 'Unable to load available vehicles');
    } finally {
      setIsCarsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadCars();
  }, [loadCars]);

  useEffect(() => {
    const storedAuth = readStoredAuth();
    if (!storedAuth?.token) {
      return;
    }

    let isMounted = true;
    setIsSessionLoading(true);

    getCurrentUser(storedAuth.token)
      .then((currentUser) => {
        if (!isMounted) {
          return;
        }

        const nextAuth = {
          token: storedAuth.token,
          user: currentUser,
        };
        saveAuth(nextAuth);
        setAuth(nextAuth);
      })
      .catch(() => {
        if (!isMounted) {
          return;
        }
        clearStoredAuth();
        setAuth(null);
      })
      .finally(() => {
        if (isMounted) {
          setIsSessionLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, []);

  useEffect(() => {
    if (activePage === 'admin' && user?.role !== 'admin') {
      setActivePage('home');
    }
  }, [activePage, user?.role]);

  const handleNavigate = (page: PageKey) => {
    setActivePage(page);
    setIsMenuOpen(false);
  };

  const handleAuthenticated = (nextAuth: AuthResponse) => {
    saveAuth(nextAuth);
    setAuth(nextAuth);
  };

  const handleLogout = () => {
    clearStoredAuth();
    setAuth(null);
    if (activePage === 'admin') {
      setActivePage('home');
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-black to-gray-800 text-white overflow-x-hidden">
      <Navbar
        pages={navPages}
        activePage={activePage}
        isMenuOpen={isMenuOpen}
        onToggleMenu={() => setIsMenuOpen((current) => !current)}
        onNavigate={handleNavigate}
        actions={
          <AuthMenu
            user={user}
            isSessionLoading={isSessionLoading}
            onAuthenticated={handleAuthenticated}
            onLogout={handleLogout}
            onAdminClick={() => handleNavigate('admin')}
          />
        }
      />

      {isAdminPage && auth?.token && user?.role === 'admin' ? (
        <AdminDashboard token={auth.token} onInventoryChanged={loadCars} onUnauthorized={handleLogout} />
      ) : (
        <>
          {isHome && (
            <HeroSection
              content={pageContent.home}
              buttonText="View Showroom"
              onButtonClick={() => handleNavigate('showroom')}
              isHome
            />
          )}

          <div className={isHome ? '' : 'pt-24'}>
            {(isHome || activePage === 'services') && <ServicesSection items={services} />}

            {(isHome || activePage === 'showroom') && <AICarAssistant />}

            {(isHome || activePage === 'showroom') && (
              <ShowroomSection cars={cars} isLoading={isCarsLoading} error={carsError} onRetry={loadCars} />
            )}

            {(isHome || activePage === 'how-it-works') && <HowItWorksSection items={howItWorksSteps} />}

            {(isHome || activePage === 'why-choose-us') && <WhyChooseUsSection items={benefits} />}

            {activePage === 'contact' && <ContactSection contact={contactInfo} />}

            {isHome && <CtaSection />}
          </div>
        </>
      )}

      <Footer contact={contactInfo} />
    </div>
  );
};

export default App;
