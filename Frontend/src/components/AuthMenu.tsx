import { LogIn, LogOut, LayoutDashboard, UserCircle, UserPlus, X } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { login, registerUser } from '../lib/api';
import type { AuthResponse, User } from '../types/api';

type AuthMode = 'login' | 'register';

interface AuthMenuProps {
  user: User | null;
  isSessionLoading: boolean;
  onAuthenticated: (auth: AuthResponse) => void;
  onLogout: () => void;
  onAdminClick: () => void;
}

const inputClass =
  'w-full bg-black/70 border border-cyan-500/30 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:border-cyan-400 focus:outline-none transition-colors';

const tabClass = (active: boolean) =>
  `flex-1 rounded-lg px-3 py-2 text-sm font-semibold transition-colors ${
    active ? 'bg-cyan-500 text-black' : 'text-gray-300 hover:bg-white/10'
  }`;

const AuthMenu = ({ user, isSessionLoading, onAuthenticated, onLogout, onAdminClick }: AuthMenuProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [mode, setMode] = useState<AuthMode>('login');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

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
    <div className="relative">
      <button
        type="button"
        onClick={() => setIsOpen((current) => !current)}
        className="inline-flex h-10 items-center gap-2 rounded-lg border border-cyan-500/30 bg-black/50 px-3 text-sm font-semibold text-cyan-100 hover:border-cyan-300/60 hover:bg-cyan-500/10 transition-colors"
        aria-expanded={isOpen}
      >
        <UserCircle size={18} />
        <span className="hidden sm:inline">{isSessionLoading ? 'Checking...' : user?.name || 'Account'}</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-3 w-[min(21rem,calc(100vw-2rem))] rounded-xl border border-cyan-500/20 bg-gray-950/95 p-4 shadow-2xl shadow-black/40 backdrop-blur-xl">
          <div className="mb-4 flex items-center justify-between gap-3">
            <p className="text-sm font-semibold text-cyan-100">{user ? 'Account' : mode === 'login' ? 'Login' : 'Register'}</p>
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
                <p className="text-sm text-gray-400">Signed in as</p>
                <p className="font-semibold text-white">{user.name}</p>
                <p className="text-sm text-gray-400">{user.email}</p>
                <span className="mt-2 inline-flex rounded-full border border-cyan-300/30 bg-cyan-500/10 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-cyan-200">
                  {user.role}
                </span>
              </div>

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
                  Admin Dashboard
                </button>
              )}

              <button
                type="button"
                onClick={() => {
                  onLogout();
                  setIsOpen(false);
                }}
                className="flex w-full items-center justify-center gap-2 rounded-lg border border-red-300/30 px-4 py-2 text-sm font-semibold text-red-200 hover:bg-red-500/10 transition-colors"
              >
                <LogOut size={16} />
                Sign Out
              </button>
            </div>
          ) : (
            <form className="space-y-4" onSubmit={handleSubmit}>
              <div className="flex rounded-xl bg-black/40 p-1">
                <button type="button" onClick={() => handleModeChange('login')} className={tabClass(mode === 'login')}>
                  Login
                </button>
                <button type="button" onClick={() => handleModeChange('register')} className={tabClass(mode === 'register')}>
                  Register
                </button>
              </div>

              {mode === 'register' && (
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  <label className="block space-y-2 text-sm text-gray-300">
                    <span>First Name</span>
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
                    <span>Last Name</span>
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
                <span>Email</span>
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
                <span>Password</span>
                <input
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  required
                  minLength={mode === 'register' ? 8 : undefined}
                  maxLength={72}
                  className={inputClass}
                  placeholder={mode === 'register' ? 'At least 8 characters' : 'Your password'}
                />
              </label>

              {mode === 'register' && (
                <label className="block space-y-2 text-sm text-gray-300">
                  <span>Confirm Password</span>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(event) => setConfirmPassword(event.target.value)}
                    required
                    minLength={8}
                    maxLength={72}
                    className={inputClass}
                    placeholder="Repeat your password"
                  />
                </label>
              )}

              {error && (
                <p className="rounded-lg border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200" role="alert">
                  {error}
                </p>
              )}

              <button
                type="submit"
                disabled={isSubmitting}
                className="flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-4 py-2 text-sm font-semibold text-black hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60 transition-colors"
              >
                {mode === 'login' ? <LogIn size={16} /> : <UserPlus size={16} />}
                {isSubmitting ? 'Please wait...' : mode === 'login' ? 'Login' : 'Create Account'}
              </button>
            </form>
          )}
        </div>
      )}
    </div>
  );
};

export default AuthMenu;
