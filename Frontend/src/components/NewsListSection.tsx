import { motion } from 'framer-motion';
import { Calendar, Newspaper, RefreshCw, X } from 'lucide-react';
import { useMemo, useState } from 'react';
import type { NewsArticle } from '../types/api';

interface NewsListSectionProps {
  articles: NewsArticle[];
  error: string;
  isLoading: boolean;
  onRetry: () => void;
}

const fallbackImageUrl = `${import.meta.env.BASE_URL}hero-main.png`;

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
});

const formatDate = (value?: string) => {
  if (!value) {
    return 'Draft date';
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Date unavailable' : dateFormatter.format(date);
};

const NewsListSection = ({ articles, error, isLoading, onRetry }: NewsListSectionProps) => {
  const [selectedArticle, setSelectedArticle] = useState<NewsArticle | null>(null);
  const featuredArticle = articles[0] ?? null;
  const remainingArticles = useMemo(() => articles.slice(1), [articles]);

  return (
    <section className="px-4 py-16">
      <div className="mx-auto max-w-7xl">
        <div className="mb-10 flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div>
            <p className="inline-flex items-center gap-2 text-sm font-semibold uppercase tracking-[0.24em] text-cyan-300">
              <Newspaper size={17} />
              News list
            </p>
            <h2 className="mt-2 text-4xl font-bold text-white">AutoRent News</h2>
            <p className="mt-3 max-w-2xl text-gray-300">Latest announcements and fleet updates published by the admin team.</p>
          </div>
          <button
            type="button"
            onClick={onRetry}
            disabled={isLoading}
            className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-4 py-3 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-60"
          >
            <RefreshCw size={16} className={isLoading ? 'animate-spin' : ''} />
            Refresh
          </button>
        </div>

        {error ? (
          <div className="rounded-xl border border-red-400/30 bg-red-500/10 px-5 py-8 text-center text-red-100">
            <p className="text-lg font-semibold">Unable to load news.</p>
            <p className="mt-2 text-sm text-red-100/80">{error}</p>
          </div>
        ) : isLoading ? (
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1.15fr_0.85fr]" aria-label="Loading news">
            <div className="h-[28rem] animate-pulse rounded-xl border border-cyan-500/10 bg-black/40" />
            <div className="space-y-4">
              {[0, 1, 2].map((item) => (
                <div key={item} className="h-36 animate-pulse rounded-xl border border-cyan-500/10 bg-black/40" />
              ))}
            </div>
          </div>
        ) : articles.length === 0 ? (
          <div className="rounded-xl border border-cyan-500/10 bg-black/35 px-5 py-12 text-center text-gray-300">
            <p className="text-lg font-semibold text-white">No published news yet.</p>
            <p className="mt-2 text-sm text-gray-400">Published admin posts will appear here.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1.15fr_0.85fr]">
            {featuredArticle && (
              <motion.article
                initial={{ opacity: 0, y: 24 }}
                whileInView={{ opacity: 1, y: 0 }}
                className="overflow-hidden rounded-xl border border-cyan-500/20 bg-black/45 shadow-lg shadow-black/25"
              >
                <div className="relative h-64 md:h-80">
                  <img
                    src={featuredArticle.image_url || fallbackImageUrl}
                    alt={featuredArticle.title}
                    referrerPolicy="no-referrer"
                    className="h-full w-full object-cover"
                    onError={(event) => {
                      event.currentTarget.onerror = null;
                      event.currentTarget.src = fallbackImageUrl;
                    }}
                  />
                  <div className="absolute inset-0 bg-gradient-to-t from-black/85 to-transparent" />
                  <span className="absolute left-4 top-4 rounded-full border border-cyan-300/30 bg-cyan-500/20 px-3 py-1 text-xs font-semibold text-cyan-100">
                    Featured
                  </span>
                </div>
                <div className="p-6">
                  <p className="inline-flex items-center gap-2 text-sm text-gray-400">
                    <Calendar size={16} className="text-cyan-300" />
                    {formatDate(featuredArticle.published_at || featuredArticle.created_at)}
                  </p>
                  <h3 className="mt-3 text-3xl font-bold text-white">{featuredArticle.title}</h3>
                  <p className="mt-3 text-gray-300">{featuredArticle.summary}</p>
                  <button
                    type="button"
                    onClick={() => setSelectedArticle(featuredArticle)}
                    className="mt-5 rounded-lg bg-cyan-500 px-5 py-3 text-sm font-semibold text-black transition-colors hover:bg-cyan-400"
                  >
                    Read more
                  </button>
                </div>
              </motion.article>
            )}

            <div className="space-y-4">
              {remainingArticles.length === 0 ? (
                <div className="rounded-xl border border-cyan-500/10 bg-black/35 p-6 text-gray-300">
                  More published news will appear here.
                </div>
              ) : (
                remainingArticles.map((article, index) => (
                  <motion.article
                    key={article.id}
                    initial={{ opacity: 0, y: 18 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.04 }}
                    className="grid gap-4 rounded-xl border border-cyan-500/20 bg-black/45 p-4 shadow-lg shadow-black/20 sm:grid-cols-[9rem_1fr]"
                  >
                    <img
                      src={article.image_url || fallbackImageUrl}
                      alt={article.title}
                      referrerPolicy="no-referrer"
                      className="h-36 w-full rounded-lg object-cover sm:h-full"
                      onError={(event) => {
                        event.currentTarget.onerror = null;
                        event.currentTarget.src = fallbackImageUrl;
                      }}
                    />
                    <div className="min-w-0">
                      <p className="inline-flex items-center gap-2 text-xs text-gray-400">
                        <Calendar size={14} className="text-cyan-300" />
                        {formatDate(article.published_at || article.created_at)}
                      </p>
                      <h3 className="mt-2 text-xl font-semibold text-white">{article.title}</h3>
                      <p className="mt-2 line-clamp-2 text-sm text-gray-300">{article.summary}</p>
                      <button
                        type="button"
                        onClick={() => setSelectedArticle(article)}
                        className="mt-4 rounded-lg border border-cyan-500/30 px-4 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                      >
                        Read more
                      </button>
                    </div>
                  </motion.article>
                ))
              )}
            </div>
          </div>
        )}
      </div>

      {selectedArticle && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/75 px-4 py-6 backdrop-blur-sm">
          <motion.article
            initial={{ opacity: 0, y: 24, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            className="max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl border border-cyan-500/30 bg-gray-950 shadow-2xl"
          >
            <div className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-cyan-500/20 bg-gray-950/95 p-5 backdrop-blur">
              <div className="min-w-0">
                <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">News</p>
                <h3 className="mt-1 text-2xl font-bold text-white">{selectedArticle.title}</h3>
              </div>
              <button
                type="button"
                onClick={() => setSelectedArticle(null)}
                className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-100 transition-colors hover:bg-cyan-500/10"
                aria-label="Close news article"
              >
                <X size={20} />
              </button>
            </div>
            <img
              src={selectedArticle.image_url || fallbackImageUrl}
              alt={selectedArticle.title}
              referrerPolicy="no-referrer"
              className="h-64 w-full object-cover"
              onError={(event) => {
                event.currentTarget.onerror = null;
                event.currentTarget.src = fallbackImageUrl;
              }}
            />
            <div className="space-y-4 p-6">
              <p className="inline-flex items-center gap-2 text-sm text-gray-400">
                <Calendar size={16} className="text-cyan-300" />
                {formatDate(selectedArticle.published_at || selectedArticle.created_at)}
              </p>
              <p className="text-lg text-gray-200">{selectedArticle.summary}</p>
              <p className="whitespace-pre-line text-sm leading-7 text-gray-300">{selectedArticle.content}</p>
            </div>
          </motion.article>
        </div>
      )}
    </section>
  );
};

export default NewsListSection;
