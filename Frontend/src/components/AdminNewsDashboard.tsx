import { Edit3, Loader2, Newspaper, Plus, RefreshCw, Save, Trash2, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState, type ChangeEvent, type FormEvent } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { ApiError, createAdminNews, deleteAdminNews, listAdminNews, updateAdminNews } from '../lib/api';
import type { NewsArticle, NewsInput, NewsStatus } from '../types/api';

interface AdminNewsDashboardProps {
  token: string;
  onNewsChanged: () => void;
  onUnauthorized: () => void;
}

interface NewsFormState {
  title: string;
  summary: string;
  content: string;
  imageUrl: string;
  status: NewsStatus;
}

const emptyNewsForm: NewsFormState = {
  title: '',
  summary: '',
  content: '',
  imageUrl: '',
  status: 'published',
};

const inputClass =
  'w-full rounded-lg border border-cyan-500/25 bg-black/60 px-3 py-2 text-sm text-white placeholder-gray-500 transition-colors focus:border-cyan-400 focus:outline-none';

const labelClass = 'block space-y-2 text-sm text-gray-300';

const fallbackImageUrl = `${import.meta.env.BASE_URL}hero-main.png`;

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
});

const formatDate = (value?: string) => {
  if (!value) {
    return 'Not published';
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Not available' : dateFormatter.format(date);
};

const toNewsInput = (form: NewsFormState): NewsInput => {
  const imageUrl = form.imageUrl.trim();

  return {
    title: form.title.trim(),
    summary: form.summary.trim(),
    content: form.content.trim(),
    status: form.status,
    ...(imageUrl ? { image_url: imageUrl } : {}),
  };
};

const formFromArticle = (article: NewsArticle): NewsFormState => ({
  title: article.title,
  summary: article.summary,
  content: article.content,
  imageUrl: article.image_url || '',
  status: article.status === 'draft' ? 'draft' : 'published',
});

const AdminNewsDashboard = ({ token, onNewsChanged, onUnauthorized }: AdminNewsDashboardProps) => {
  const [articles, setArticles] = useState<NewsArticle[]>([]);
  const [form, setForm] = useState<NewsFormState>(emptyNewsForm);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const articleTexts = useMemo(() => articles.flatMap((article) => [article.title, article.summary].filter(Boolean)), [articles]);
  const { t } = useTranslation([
    'Not published',
    'Not available',
    'Unable to load news',
    'News article updated.',
    'News article published.',
    'News draft saved.',
    'Unable to save news',
    'Delete news article',
    'News article deleted.',
    'Unable to delete news',
    'Total Articles',
    'Published',
    'Drafts',
    'News Dashboard',
    'Edit News',
    'Publish News',
    'Create public AutoRent updates with optional cover image links.',
    'Cancel editing news',
    'Title',
    'Summary',
    'Content',
    'Cover Image URL',
    'Optional',
    'Status',
    'Draft',
    'Saving...',
    'Save News',
    'News Articles',
    'Published articles are shown on the public News list.',
    'Refresh',
    'Loading news',
    'No news yet.',
    'Use the form to publish the first update.',
    'Updated',
    'Edit',
    'Deleting...',
    'Delete',
    'published',
    'draft',
    error,
    message,
    ...articles.map((article) => article.status),
    ...articleTexts,
  ]);
  const displayDate = (value?: string) => {
    const formattedDate = formatDate(value);
    return formattedDate === 'Not published' || formattedDate === 'Not available' ? t(formattedDate) : formattedDate;
  };

  const loadNews = useCallback(async () => {
    setIsLoading(true);
    setError('');

    try {
      const loadedArticles = await listAdminNews(token);
      setArticles(loadedArticles);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Unable to load news');
      if (loadError instanceof ApiError && loadError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsLoading(false);
    }
  }, [onUnauthorized, token]);

  useEffect(() => {
    loadNews();
  }, [loadNews]);

  const stats = useMemo(() => {
    const published = articles.filter((article) => article.status === 'published').length;

    return {
      total: articles.length,
      published,
      drafts: articles.length - published,
    };
  }, [articles]);

  const updateField =
    (field: keyof NewsFormState) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
      const value = field === 'status' ? (event.target.value as NewsStatus) : event.target.value;
      setForm((current) => ({
        ...current,
        [field]: value,
      }));
    };

  const resetForm = () => {
    setForm(emptyNewsForm);
    setEditingId(null);
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSaving(true);
    setError('');
    setMessage('');

    try {
      const payload = toNewsInput(form);
      if (editingId) {
        await updateAdminNews(token, editingId, payload);
        setMessage('News article updated.');
      } else {
        await createAdminNews(token, payload);
        setMessage(payload.status === 'published' ? 'News article published.' : 'News draft saved.');
      }

      resetForm();
      await loadNews();
      onNewsChanged();
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : 'Unable to save news');
      if (saveError instanceof ApiError && saveError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleEdit = (article: NewsArticle) => {
    setForm(formFromArticle(article));
    setEditingId(article.id);
    setError('');
    setMessage('');
    document.getElementById('news-dashboard-form')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  const handleDelete = async (article: NewsArticle) => {
    const shouldDelete = window.confirm(`${t('Delete news article')} "${article.title}"?`);
    if (!shouldDelete) {
      return;
    }

    setError('');
    setMessage('');
    setDeletingId(article.id);

    try {
      await deleteAdminNews(token, article.id);
      setMessage('News article deleted.');
      if (editingId === article.id) {
        resetForm();
      }
      await loadNews();
      onNewsChanged();
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : 'Unable to delete news');
      if (deleteError instanceof ApiError && deleteError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setDeletingId(null);
    }
  };

  return (
    <section className="space-y-6">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
          <p className="text-sm text-gray-400">{t('Total Articles')}</p>
          <p className="mt-2 text-3xl font-bold text-white">{stats.total}</p>
        </div>
        <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
          <p className="text-sm text-gray-400">{t('Published')}</p>
          <p className="mt-2 text-3xl font-bold text-cyan-300">{stats.published}</p>
        </div>
        <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
          <p className="text-sm text-gray-400">{t('Drafts')}</p>
          <p className="mt-2 text-3xl font-bold text-white">{stats.drafts}</p>
        </div>
      </div>

      {(error || message) && (
        <div
          className={`rounded-xl border px-4 py-3 text-sm ${
            error ? 'border-red-400/30 bg-red-500/10 text-red-100' : 'border-cyan-400/30 bg-cyan-500/10 text-cyan-100'
          }`}
          role={error ? 'alert' : 'status'}
        >
          {t(error || message)}
        </div>
      )}

      <div className="grid grid-cols-1 gap-8 xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)]">
        <form id="news-dashboard-form" onSubmit={handleSubmit} className="rounded-xl border border-cyan-500/20 bg-white/10 p-6">
          <div className="mb-6 flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('News Dashboard')}</p>
              <h2 className="mt-1 text-2xl font-semibold text-white">{editingId ? t('Edit News') : t('Publish News')}</h2>
              <p className="mt-1 text-sm text-gray-400">{t('Create public AutoRent updates with optional cover image links.')}</p>
            </div>
            {editingId && (
              <button
                type="button"
                onClick={resetForm}
                className="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-cyan-500/20 text-cyan-100 transition-colors hover:bg-cyan-500/10"
                aria-label={t('Cancel editing news')}
              >
                <X size={18} />
              </button>
            )}
          </div>

          <div className="space-y-4">
            <label className={labelClass}>
              <span>{t('Title')}</span>
              <input value={form.title} onChange={updateField('title')} className={inputClass} required maxLength={120} />
            </label>
            <label className={labelClass}>
              <span>{t('Summary')}</span>
              <textarea
                value={form.summary}
                onChange={updateField('summary')}
                className={`${inputClass} min-h-24 resize-y`}
                required
                maxLength={240}
              />
            </label>
            <label className={labelClass}>
              <span>{t('Content')}</span>
              <textarea
                value={form.content}
                onChange={updateField('content')}
                className={`${inputClass} min-h-44 resize-y`}
                required
              />
            </label>
            <label className={labelClass}>
              <span>{t('Cover Image URL')}</span>
              <input value={form.imageUrl} onChange={updateField('imageUrl')} className={inputClass} placeholder={t('Optional')} maxLength={255} />
            </label>
            <label className={labelClass}>
              <span>{t('Status')}</span>
              <select value={form.status} onChange={updateField('status')} className={inputClass}>
                <option value="published" className="bg-gray-950">
                  {t('Published')}
                </option>
                <option value="draft" className="bg-gray-950">
                  {t('Draft')}
                </option>
              </select>
            </label>
          </div>

          <button
            type="submit"
            disabled={isSaving}
            className="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-4 py-3 text-sm font-semibold text-black transition-colors hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isSaving ? <Loader2 size={17} className="animate-spin" /> : editingId ? <Save size={17} /> : <Plus size={17} />}
            {isSaving ? t('Saving...') : editingId ? t('Save News') : t('Publish News')}
          </button>
        </form>

        <section className="rounded-xl border border-cyan-500/20 bg-white/10 p-6">
          <div className="mb-6 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
            <div>
              <h2 className="text-2xl font-semibold text-white">{t('News Articles')}</h2>
              <p className="mt-1 text-sm text-gray-400">{t('Published articles are shown on the public News list.')}</p>
            </div>
            <button
              type="button"
              onClick={loadNews}
              disabled={isLoading}
              className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-60"
            >
              <RefreshCw size={15} className={isLoading ? 'animate-spin' : ''} />
              {t('Refresh')}
            </button>
          </div>

          {isLoading ? (
            <div className="space-y-3" aria-label={t('Loading news')}>
              {[0, 1, 2].map((item) => (
                <div key={item} className="h-28 animate-pulse rounded-xl bg-black/40" />
              ))}
            </div>
          ) : articles.length === 0 ? (
            <div className="rounded-xl border border-cyan-500/10 bg-black/40 py-12 text-center">
              <Newspaper size={32} className="mx-auto text-cyan-300" />
              <p className="mt-3 text-lg font-semibold text-white">{t('No news yet.')}</p>
              <p className="mt-2 text-sm text-gray-400">{t('Use the form to publish the first update.')}</p>
            </div>
          ) : (
            <div className="space-y-4">
              {articles.map((article) => (
                <article key={article.id} className="grid gap-4 rounded-xl border border-cyan-500/10 bg-black/35 p-4 lg:grid-cols-[8rem_1fr_auto]">
                  <img
                    src={article.image_url || fallbackImageUrl}
                    alt={t(article.title)}
                    referrerPolicy="no-referrer"
                    className="h-32 w-full rounded-lg object-cover lg:h-24 lg:w-32"
                    onError={(event) => {
                      event.currentTarget.onerror = null;
                      event.currentTarget.src = fallbackImageUrl;
                    }}
                  />
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <h3 className="break-words text-lg font-semibold text-white">{t(article.title)}</h3>
                      <span className="rounded-full border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-xs capitalize text-cyan-100">
                        {t(article.status)}
                      </span>
                    </div>
                    <p className="mt-2 text-sm text-gray-300">{t(article.summary)}</p>
                    <p className="mt-2 text-xs text-gray-500">
                      {t('Published')} {displayDate(article.published_at)} | {t('Updated')} {displayDate(article.updated_at)}
                    </p>
                  </div>
                  <div className="flex items-center gap-2 lg:flex-col lg:items-stretch">
                    <button
                      type="button"
                      onClick={() => handleEdit(article)}
                      disabled={deletingId === article.id}
                      className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10"
                    >
                      <Edit3 size={16} />
                      {t('Edit')}
                    </button>
                    <button
                      type="button"
                      onClick={() => handleDelete(article)}
                      disabled={deletingId === article.id}
                      className="inline-flex items-center justify-center gap-2 rounded-lg border border-red-400/30 px-3 py-2 text-sm font-semibold text-red-200 transition-colors hover:bg-red-500/10 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      {deletingId === article.id ? <Loader2 size={16} className="animate-spin" /> : <Trash2 size={16} />}
                      {deletingId === article.id ? t('Deleting...') : t('Delete')}
                    </button>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </div>
    </section>
  );
};

export default AdminNewsDashboard;
