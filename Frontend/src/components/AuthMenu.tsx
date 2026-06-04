import { LogIn, LogOut, LayoutDashboard, UserCircle, UserPlus, UserRound, X } from 'lucide-react';
import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { googleLogin, login, registerUser } from '../lib/api';
import type { AuthResponse, User } from '../types/api';

type AuthMode = 'login' | 'register';

interface AuthMenuProps {
  user: User | null;
  isSessionLoading: boolean;
  onAuthenticated: (auth: AuthResponse) => void;
  onLogout: () => void;
  onAdminClick: () => void;
  onProfileClick: () => void;
}

const inputClass =
  'w-full bg-black/70 border border-cyan-500/30 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:border-cyan-400 focus:outline-none transition-colors';

const tabClass = (active: boolean) =>
  `flex-1 rounded-lg px-3 py-2 text-sm font-semibold transition-colors ${
    active ? 'bg-cyan-500 text-black' : 'text-gray-300 hover:bg-white/10'
  }`;

const googleClientId = (import.meta.env.VITE_GOOGLE_AUTH_CLIENT_ID || '').trim();
const googleScriptID = 'google-identity-services';
const googleScriptSrc = 'https://accounts.google.com/gsi/client';

type GoogleCredentialResponse = {
  credential?: string;
};

type GoogleIdentityAPI = {
  initialize: (config: { client_id: string; callback: (response: GoogleCredentialResponse) => void }) => void;
  renderButton: (parent: HTMLElement, options: Record<string, string | number>) => void;
  disableAutoSelect?: () => void;
};

type GoogleWindow = Window &
  typeof globalThis & {
    google?: {
      accounts?: {
        id?: GoogleIdentityAPI;
      };
    };
  };

let googleScriptPromise: Promise<void> | null = null;

const googleIdentityAPI = () => (window as GoogleWindow).google?.accounts?.id;

const loadGoogleIdentityScript = () => {
  if (typeof window === 'undefined' || typeof document === 'undefined') {
    return Promise.reject(new Error('Google login is unavailable.'));
  }
  if (googleIdentityAPI()) {
    return Promise.resolve();
  }
  if (googleScriptPromise) {
    return googleScriptPromise;
  }

  googleScriptPromise = new Promise((resolve, reject) => {
    const existingScript = document.getElementById(googleScriptID) as HTMLScriptElement | null;
    if (existingScript) {
      existingScript.addEventListener('load', () => resolve(), { once: true });
      existingScript.addEventListener('error', () => reject(new Error('Google login is unavailable.')), { once: true });
      return;
    }

    const script = document.createElement('script');
    script.id = googleScriptID;
    script.src = googleScriptSrc;
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('Google login is unavailable.'));
    document.head.appendChild(script);
  });

  return googleScriptPromise;
};

const AuthMenu = ({ user, isSessionLoading, onAuthenticated, onLogout, onAdminClick, onProfileClick }: AuthMenuProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [mode, setMode] = useState<AuthMode>('login');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isGoogleSubmitting, setIsGoogleSubmitting] = useState(false);
  const googleButtonRef = useRef<HTMLDivElement | null>(null);
  const { t } = useTranslation([
    'Checking...',
    'Account',
    'Login',
    'Register',
    'Signed in as',
    'View Profile',
    'Admin Dashboard',
    'Sign Out',
    'First Name',
    'Last Name',
    'Email',
    'Password',
    'Confirm Password',
    'At least 8 characters',
    'Your password',
    'Repeat your password',
    'Passwords do not match',
    'Unable to complete request',
    'Google login is unavailable.',
    'Google login is not configured.',
    'or',
    'Please wait...',
    'Create Account',
    user?.role,
    error,
  ]);

  const resetForm = () => {
    setFirstName('');
    setLastName('');
    setEmail('');
    setPassword('');
    setConfirmPassword('');
    setError('');
  };

  const handleModeChange = (nextMode: AuthMode) => {
    setMode(nextMode);
    setError('');
  };

  useEffect(() => {
    if (!isOpen || user || !googleClientId || !googleButtonRef.current) {
      return;
    }

    let isCancelled = false;
    setError('');

    loadGoogleIdentityScript()
      .then(() => {
        if (isCancelled || !googleButtonRef.current) {
          return;
        }

        const identityAPI = googleIdentityAPI();
        if (!identityAPI) {
          throw new Error('Google login is unavailable.');
        }

        identityAPI.initialize({
          client_id: googleClientId,
          callback: async (response) => {
            if (!response.credential) {
              setError('Google login is unavailable.');
              return;
            }

            setError('');
            setIsGoogleSubmitting(true);
            try {
              const auth = await googleLogin({ credential: response.credential });
              onAuthenticated(auth);
              resetForm();
              setIsOpen(false);
            } catch (submitError) {
              setError(submitError instanceof Error ? submitError.message : 'Unable to complete request');
            } finally {
              setIsGoogleSubmitting(false);
            }
          },
        });

        googleButtonRef.current.innerHTML = '';
        identityAPI.renderButton(googleButtonRef.current, {
          theme: 'filled_black',
          size: 'large',
          type: 'standard',
          shape: 'rectangular',
          text: mode === 'register' ? 'signup_with' : 'signin_with',
          logo_alignment: 'left',
          width: 304,
        });
      })
      .catch((loadError) => {
        if (!isCancelled) {
          setError(loadError instanceof Error ? loadError.message : 'Google login is unavailable.');
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [isOpen, mode, onAuthenticated, user]);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');

    if (mode === 'register' && password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    setIsSubmitting(true);

    try {
      const trimmedFirstName = firstName.trim();
      const trimmedLastName = lastName.trim();
      const auth =
        mode === 'login'
          ? await login({ email, password })
          : await registerUser({
              first_name: trimmedFirstName,
              last_name: trimmedLastName,
              email,
              password,
            });

      onAuthenticated(auth);
      resetForm();
      setIsOpen(false);
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Unable to complete request');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="relative shrink-0">
      <button
        type="button"
        onClick={() => setIsOpen((current) => !current)}
        className="inline-flex h-10 max-w-44 items-center gap-2 rounded-lg border border-cyan-500/30 bg-black/50 px-3 text-sm font-semibold text-cyan-100 transition-colors hover:border-cyan-300/60 hover:bg-cyan-500/10 sm:max-w-52"
        aria-expanded={isOpen}
      >
        <UserCircle size={18} className="shrink-0" />
        <span className="hidden min-w-0 truncate whitespace-nowrap sm:inline">{isSessionLoading ? t('Checking...') : user?.name || t('Account')}</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-3 w-[min(21rem,calc(100vw-2rem))] rounded-xl border border-cyan-500/20 bg-gray-950/95 p-4 shadow-2xl shadow-black/40 backdrop-blur-xl">
          <div className="mb-4 flex items-center justify-between gap-3">
            <p className="text-sm font-semibold text-cyan-100">{user ? t('Account') : mode === 'login' ? t('Login') : t('Register')}</p>
            <button
              type="button"
              onClick={() => setIsOpen(false)}
              className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-cyan-500/20 text-cyan-100 hover:bg-cyan-500/10 transition-colors"
              aria-label="Close account panel"
            >
              <X size={17} />
            </button>
          </div>

          {user ? (
            <div className="space-y-4">
              <div>
                <p className="text-sm text-gray-400">{t('Signed in as')}</p>
                <p className="font-semibold text-white">{user.name}</p>
                <p className="text-sm text-gray-400">{user.email}</p>
                <span className="mt-2 inline-flex rounded-full border border-cyan-300/30 bg-cyan-500/10 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-cyan-200">
                  {t(user.role)}
                </span>
              </div>

              <button
                type="button"
                onClick={() => {
                  onProfileClick();
                  setIsOpen(false);
                }}
                className="flex w-full items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-4 py-2 text-sm font-semibold text-cyan-100 hover:bg-cyan-500/10 transition-colors"
              >
                <UserRound size={16} />
                {t('View Profile')}
              </button>

              {user.role === 'admin' && (
                <button
                  type="button"
                  onClick={() => {
                    onAdminClick();
                    setIsOpen(false);
                  }}
                  className="flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-4 py-2 text-sm font-semibold text-black hover:bg-cyan-400 transition-colors"
                >
                  <LayoutDashboard size={16} />
                  {t('Admin Dashboard')}
                </button>
              )}

              <button
                type="button"
                onClick={() => {
                  googleIdentityAPI()?.disableAutoSelect?.();
                  onLogout();
                  setIsOpen(false);
                }}
                className="flex w-full items-center justify-center gap-2 rounded-lg border border-red-300/30 px-4 py-2 text-sm font-semibold text-red-200 hover:bg-red-500/10 transition-colors"
              >
                <LogOut size={16} />
                {t('Sign Out')}
              </button>
            </div>
          ) : (
            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="flex rounded-xl bg-black/40 p-1">
                <button type="button" onClick={() => handleModeChange('login')} className={tabClass(mode === 'login')}>
                  {t('Login')}
                </button>
                <button type="button" onClick={() => handleModeChange('register')} className={tabClass(mode === 'register')}>
                  {t('Register')}
                </button>
              </div>

              {mode === 'register' && (
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <label className="block space-y-2 text-sm text-gray-300">
                    <span>{t('First Name')}</span>
                    <input
                      type="text"
                      value={firstName}
                      onChange={(event) => setFirstName(event.target.value)}
                      required
                      maxLength={50}
                      className={inputClass}
                      placeholder="Jane"
                    />
                  </label>

                  <label className="block space-y-2 text-sm text-gray-300">
                    <span>{t('Last Name')}</span>
                    <input
                      type="text"
                      value={lastName}
                      onChange={(event) => setLastName(event.target.value)}
                      required
                      maxLength={50}
                      className={inputClass}
                      placeholder="Driver"
                    />
                  </label>
                </div>
              )}

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Email')}</span>
                <input
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  required
                  maxLength={255}
                  className={inputClass}
                  placeholder="you@example.com"
                />
              </label>

              <label className="block space-y-2 text-sm text-gray-300">
                <span>{t('Password')}</span>
                <input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  required
                  minLength={mode === 'register' ? 8 : undefined}
                  maxLength={72}
                  className={inputClass}
                  placeholder={mode === 'register' ? t('At least 8 characters') : t('Your password')}
                />
              </label>

              {mode === 'register' && (
                <label className="block space-y-2 text-sm text-gray-300">
                    <span>{t('Confirm Password')}</span>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(event) => setConfirmPassword(event.target.value)}
                    required
                    minLength={8}
                    maxLength={72}
                    className={inputClass}
                    placeholder={t('Repeat your password')}
                  />
                </label>
              )}

              {error && (
                <p className="rounded-lg border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200" role="alert">
                  {t(error)}
                </p>
              )}

              <button
                type="submit"
                disabled={isSubmitting || isGoogleSubmitting}
                className="flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-4 py-2 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
              >
                {mode === 'login' ? <LogIn size={16} /> : <UserPlus size={16} />}
                {isSubmitting ? t('Please wait...') : mode === 'login' ? t('Login') : t('Create Account')}
              </button>

              <div className="flex items-center gap-3 text-xs uppercase tracking-wide text-gray-500">
                <span className="h-px flex-1 bg-cyan-500/15" />
                {t('or')}
                <span className="h-px flex-1 bg-cyan-500/15" />
              </div>

              {googleClientId ? (
                <div
                  ref={googleButtonRef}
                  className={`flex min-h-11 justify-center ${isGoogleSubmitting ? 'pointer-events-none opacity-60' : ''}`}
                  aria-busy={isGoogleSubmitting}
                />
              ) : (
                <p className="rounded-lg border border-amber-300/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-100">
                  {t('Google login is not configured.')}
                </p>
              )}
            </form>
          )}
        </div>
      )}
    </div>
  );
};

export default AuthMenu;
