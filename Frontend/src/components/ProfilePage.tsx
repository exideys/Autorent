import { CalendarDays, Hash, LayoutDashboard, LogOut, Mail, Shield, Star, UserCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { ApiError, listMyRentalOrders } from '../lib/api';
import type { RentalOrder, User } from '../types/api';

interface ProfilePageProps {
  user: User;
  token: string;
  onAdminClick: () => void;
  onLogout: () => void;
  onShowroomClick: () => void;
}

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'long',
  day: 'numeric',
  year: 'numeric',
});

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
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

const ProfilePage = ({ user, token, onAdminClick, onLogout, onShowroomClick }: ProfilePageProps) => {
  const [orders, setOrders] = useState<RentalOrder[]>([]);
  const [isOrdersLoading, setIsOrdersLoading] = useState(true);
  const [ordersError, setOrdersError] = useState('');
  const displayName = user.name?.trim();
  const firstName = user.first_name?.trim();
  const lastName = user.last_name?.trim();
  const email = user.email;
  const role = user.role || 'user';
  const status = user.status || 'unknown';
  const rating = Number.isFinite(user.rating) ? user.rating.toFixed(1) : '';
  const ratingCount = Number.isFinite(user.rating_count) ? user.rating_count : 0;
  const { t } = useTranslation([
    'Not available',
    'AutoRent member',
    'Not specified',
    'Email not available',
    'Not rated',
    'Profile',
    'View Showroom',
    'Admin Dashboard',
    'Sign Out',
    'Account details',
    'Your current session profile from AutoRent authentication.',
    'ID',
    'First name',
    'Last name',
    'Email',
    'Role',
    'Status',
    'Rating',
    'Rating count',
    'Created at',
    'Updated at',
    'Rentals',
    'Your placed orders',
    'active',
    'Unable to load rental orders',
    'You do not have rental orders yet.',
    'Total',
    'Deposit',
    'to',
    'at',
    role,
    status,
    ordersError,
    ...orders.map((order) => order.status),
  ]);
  const displayDate = (value: string) => {
    const formattedDate = formatDate(value);
    return formattedDate === 'Not available' ? t('Not available') : formattedDate;
  };

  useEffect(() => {
    if (!token) {
      setOrders([]);
      setIsOrdersLoading(false);
      return;
    }

    let isMounted = true;
    setIsOrdersLoading(true);
    setOrdersError('');

    listMyRentalOrders(token)
      .then((loadedOrders) => {
        if (isMounted) {
          setOrders(loadedOrders);
        }
      })
      .catch((loadError) => {
        if (!isMounted) {
          return;
        }
        setOrdersError(loadError instanceof Error ? loadError.message : 'Unable to load rental orders');
        if (loadError instanceof ApiError && loadError.status === 401) {
          onLogout();
        }
      })
      .finally(() => {
        if (isMounted) {
          setIsOrdersLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, [onLogout, token]);

  return (
    <main className="min-h-screen px-4 py-28">
      <div className="mx-auto max-w-5xl">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(20rem,1.1fr)]">
          <section className="rounded-2xl border border-cyan-500/20 bg-black/45 p-6 shadow-2xl shadow-black/30">
            <div className="flex items-center gap-4">
              <div className="flex h-20 w-20 items-center justify-center rounded-2xl border border-cyan-300/30 bg-cyan-500/10 text-2xl font-bold text-cyan-200">
                {initialsFor(displayName || 'AutoRent member')}
              </div>
              <div className="min-w-0">
                <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Profile')}</p>
                <h1 className="mt-1 break-words text-3xl font-bold text-white">{displayName || t('AutoRent member')}</h1>
                <p className="mt-1 break-words text-sm text-gray-400">{email || t('Email not available')}</p>
              </div>
            </div>

            <div className="mt-8 grid gap-3">
              <button
                type="button"
                onClick={onShowroomClick}
                className="rounded-lg bg-cyan-500 px-4 py-3 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
              >
                {t('View Showroom')}
              </button>
              {role === 'admin' && (
                <button
                  type="button"
                  onClick={onAdminClick}
                  className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-4 py-3 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                >
                  <LayoutDashboard size={17} />
                  {t('Admin Dashboard')}
                </button>
              )}
              <button
                type="button"
                onClick={onLogout}
                className="inline-flex items-center justify-center gap-2 rounded-lg border border-red-300/30 px-4 py-3 text-sm font-semibold text-red-200 transition-colors hover:bg-red-500/10"
              >
                <LogOut size={17} />
                {t('Sign Out')}
              </button>
            </div>
          </section>

          <section className="rounded-2xl border border-cyan-500/20 bg-black/35 p-6">
            <h2 className="text-2xl font-bold text-white">{t('Account details')}</h2>
            <p className="mt-2 text-gray-400">{t('Your current session profile from AutoRent authentication.')}</p>

            <div className="mt-6 grid gap-4 sm:grid-cols-2">
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Hash size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('ID')}</p>
                <p className="mt-1 text-lg font-semibold text-white">#{user.id}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <UserCircle size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('First name')}</p>
                <p className="mt-1 break-words text-lg font-semibold text-white">{firstName || t('Not specified')}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <UserCircle size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Last name')}</p>
                <p className="mt-1 break-words text-lg font-semibold text-white">{lastName || t('Not specified')}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Mail size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Email')}</p>
                <p className="mt-1 break-words text-lg font-semibold text-white">{email || t('Email not available')}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Shield size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Role')}</p>
                <p className="mt-1 text-lg font-semibold capitalize text-white">{t(role)}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Shield size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Status')}</p>
                <p className="mt-1 text-lg font-semibold capitalize text-white">{t(status)}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Star size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Rating')}</p>
                <p className="mt-1 text-lg font-semibold text-white">{rating || t('Not rated')}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <Star size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Rating count')}</p>
                <p className="mt-1 text-lg font-semibold text-white">{ratingCount}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <CalendarDays size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Created at')}</p>
                <p className="mt-1 text-lg font-semibold text-white">{displayDate(user.created_at)}</p>
              </div>
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-4">
                <CalendarDays size={22} className="text-cyan-300" />
                <p className="mt-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{t('Updated at')}</p>
                <p className="mt-1 text-lg font-semibold text-white">{displayDate(user.updated_at)}</p>
              </div>
            </div>
          </section>
        </div>

        <section className="mt-6 rounded-2xl border border-cyan-500/20 bg-black/35 p-6">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Rentals')}</p>
              <h2 className="mt-1 text-2xl font-bold text-white">{t('Your placed orders')}</h2>
            </div>
            <span className="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-3 py-1 text-sm text-cyan-100">
              {orders.length} {t('active')}
            </span>
          </div>

          {ordersError && (
            <p className="mt-4 rounded-lg border border-red-400/30 bg-red-500/10 px-3 py-2 text-sm text-red-200" role="alert">
              {t(ordersError)}
            </p>
          )}

          {isOrdersLoading ? (
            <div className="mt-6 space-y-3" aria-label="Loading rental orders">
              {[0, 1].map((item) => (
                <div key={item} className="h-28 animate-pulse rounded-xl bg-black/40" />
              ))}
            </div>
          ) : orders.length === 0 ? (
            <div className="mt-6 rounded-xl border border-cyan-500/10 bg-black/35 px-4 py-8 text-center text-gray-300">
              {t('You do not have rental orders yet.')}
            </div>
          ) : (
            <div className="mt-6 space-y-4">
              {orders.map((order) => (
                <article key={order.id} className="grid gap-4 rounded-xl border border-cyan-500/10 bg-black/35 p-4 md:grid-cols-[1fr_auto]">
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <h3 className="text-lg font-semibold text-white">
                        {order.car.brand} {order.car.model} ({order.car.year})
                      </h3>
                      <span className="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-xs capitalize text-cyan-100">
                        {t(order.status)}
                      </span>
                    </div>
                    <p className="mt-2 text-sm text-gray-300">
                      {displayDate(order.start_date)} {t('to')} {displayDate(order.end_date)} {t('at')} {order.pickup_time}
                    </p>
                    <p className="mt-1 text-sm text-gray-400">{order.pickup_location}</p>
                    {order.notes && <p className="mt-2 text-sm text-gray-500">{order.notes}</p>}
                  </div>
                  <div className="grid grid-cols-2 gap-3 text-sm md:w-64">
                    <div className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                      <p className="text-gray-500">{t('Total')}</p>
                      <p className="mt-1 font-semibold text-cyan-300">{currencyFormatter.format(order.total_price)}</p>
                    </div>
                    <div className="rounded-lg border border-cyan-500/10 bg-black/30 p-3">
                      <p className="text-gray-500">{t('Deposit')}</p>
                      <p className="mt-1 font-semibold text-white">{currencyFormatter.format(order.deposit)}</p>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </div>
    </main>
  );
};

export default ProfilePage;
