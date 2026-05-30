import { CalendarDays, Hash, LayoutDashboard, LogOut, Mail, Shield, Star, UserCircle } from 'lucide-react';
import type { User } from '../types/api';

interface ProfilePageProps {
  user: User;
  onAdminClick: () => void;
  onLogout: () => void;
  onShowroomClick: () => void;
}

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'long',
  day: 'numeric',
  year: 'numeric',
});

const formatDate = (value: string) => {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Not available' : dateFormatter.format(date);
};

const initialsFor = (name: string) => {
  const initials = name
    .trim()
    .split(/\s+/)
    .map((part) => part[0])
    .join('')
    .slice(0, 2)
    .toUpperCase();

  return initials || 'AR';
};

const ProfilePage = ({ user, onAdminClick, onLogout, onShowroomClick }: ProfilePageProps) => {
  const displayName = user.name?.trim() || 'AutoRent member';
  const firstName = user.first_name?.trim() || 'Not specified';
  const lastName = user.last_name?.trim() || 'Not specified';
  const email = user.email || 'Email not available';
  const role = user.role || 'user';
  const status = user.status || 'unknown';
  const rating = Number.isFinite(user.rating) ? user.rating.toFixed(1) : 'Not rated';
  const ratingCount = Number.isFinite(user.rating_count) ? user.rating_count : 0;

  return (
    <main className="min-h-screen px-4 py-28">
      <div className="mx-auto max-w-5xl">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(20rem,1.1fr)]">
          <section className="rounded-2xl border border-cyan-500/20 bg-black/45 p-6 shadow-2xl shadow-black/30">
            <div className="flex items-center gap-4">
              <div className="flex h-20 w-20 items-center justify-center rounded-2xl border border-cyan-300/30 bg-cyan-500/10 text-2xl font-bold text-cyan-200">
                {initialsFor(displayName)}
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">Profile</p>
                <h1 className="mt-1 break-words text-3xl font-bold text-white">{displayName}</h1>
                <p className="mt-1 break-words text-sm text-gray-400">{email}</p>
              </div>
            </div>

            <div className="mt-8 grid gap-3">
              <button
                type="button"
                onClick={onShowroomClick}
                className="rounded-lg bg-cyan-500 px-4 py-3 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
              >
                Open Showroom
              </button>
              {role === 'admin' && (
                <button
                  type="button"
                  onClick={onAdminClick}
                  className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-4 py-3 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                >
                  <LayoutDashboard size={17} />
                  Admin Dashboard
                </button>
              )}
              <button
                type="button"
                onClick={onLogout}
                className="inline-flex items-center justify-center gap-2 rounded-lg border border-red-300/30 px-4 py-3 text-sm font-semibold text-red-200 transition-colors hover:bg-red-500/10"
              >
                <LogOut size={17} />
                Sign Out
              </button>
            </div>
          </section>

          <section className="rounded-2xl border border-cyan-500/20 bg-black/35 p-6">
            <h2 className="text-2xl font-bold text-white">Account details</h2>
            <p className="mt-2 text-gray-400">Your current session profile from AutoRent authentication.</p>

            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Hash size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">ID</p>
                <p className="mt-1 text-lg font-semibold text-white">#{user.id}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <UserCircle size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">First name</p>
                <p className="mt-1 break-words text-lg font-semibold text-white">{firstName}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <UserCircle size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Last name</p>
                <p className="mt-1 break-words text-lg font-semibold text-white">{lastName}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Mail size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Email</p>
                <p className="mt-1 break-words text-lg font-semibold text-white">{email}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Shield size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Role</p>
                <p className="mt-1 text-lg font-semibold capitalize text-white">{role}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Shield size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Status</p>
                <p className="mt-1 text-lg font-semibold capitalize text-white">{status}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Star size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Rating</p>
                <p className="mt-1 text-lg font-semibold text-white">{rating}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Star size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Rating count</p>
                <p className="mt-1 text-lg font-semibold text-white">{ratingCount}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <CalendarDays size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Created at</p>
                <p className="mt-1 text-lg font-semibold text-white">{formatDate(user.created_at)}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <CalendarDays size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Updated at</p>
                <p className="mt-1 text-lg font-semibold text-white">{formatDate(user.updated_at)}</p>
              </div>
            </div>
          </section>
        </div>
      </div>
    </main>
  );
};

export default ProfilePage;
