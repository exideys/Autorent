import { Download, Inbox, Loader2, Lock, MessageSquare, RefreshCw, Reply, Send, Unlock } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import {
  ApiError,
  downloadSupportAttachment,
  getAdminSupportConversation,
  listAdminSupportConversations,
  replyAdminSupportMessage,
  streamAdminSupportEvents,
  updateAdminSupportConversationStatus,
} from '../lib/api';
import type { SupportAttachment, SupportConversation } from '../types/api';
import { isImageAttachment, SupportImageAttachmentPreview } from './SupportImagePreview';

interface AdminSupportDashboardProps {
  token: string;
  onUnauthorized: () => void;
}

type SupportStatusFilter = 'open' | 'closed';

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
});

const reconnectDelayMs = 3000;

const formatDate = (value?: string) => {
  if (!value) {
    return 'No messages';
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Not available' : dateFormatter.format(date);
};

const formatFileSize = (size: number) => {
  if (size < 1024 * 1024) {
    return `${Math.max(1, Math.round(size / 1024))} KB`;
  }

  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
};

const lastMessageText = (conversation: SupportConversation) => {
  const messages = conversation.messages || [];
  const lastMessage = messages[messages.length - 1];
  if (!lastMessage) {
    return 'No messages';
  }
  if (lastMessage.body.trim()) {
    return lastMessage.body;
  }
  return 'Attachment';
};

const customerName = (conversation?: SupportConversation | null) =>
  conversation?.user?.name || conversation?.user?.email || `User #${conversation?.user_id || ''}`;

const isAbortError = (error: unknown) => error instanceof DOMException && error.name === 'AbortError';

const statusLabel = (status?: string) => (status === 'closed' ? 'Closed' : 'Open');

const AdminSupportDashboard = ({ token, onUnauthorized }: AdminSupportDashboardProps) => {
  const [conversations, setConversations] = useState<SupportConversation[]>([]);
  const [selectedConversation, setSelectedConversation] = useState<SupportConversation | null>(null);
  const selectedConversationRef = useRef<SupportConversation | null>(null);
  const [activeStatusFilter, setActiveStatusFilter] = useState<SupportStatusFilter>('open');
  const activeStatusFilterRef = useRef<SupportStatusFilter>('open');
  const [replyText, setReplyText] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isConversationLoading, setIsConversationLoading] = useState(false);
  const [isReplying, setIsReplying] = useState(false);
  const [isStatusUpdating, setIsStatusUpdating] = useState(false);
  const [downloadingAttachmentID, setDownloadingAttachmentID] = useState<number | null>(null);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');
  const selectedConversationID = selectedConversation?.id || null;
  const totalMessages = useMemo(
    () => conversations.reduce((total, conversation) => total + (conversation.messages?.length || 0), 0),
    [conversations],
  );
  const openConversationCount = useMemo(() => conversations.filter((conversation) => conversation.status !== 'closed').length, [conversations]);
  const closedConversationCount = useMemo(() => conversations.filter((conversation) => conversation.status === 'closed').length, [conversations]);
  const visibleConversations = useMemo(
    () =>
      conversations.filter((conversation) =>
        activeStatusFilter === 'closed' ? conversation.status === 'closed' : conversation.status !== 'closed',
      ),
    [activeStatusFilter, conversations],
  );
  const { t } = useTranslation([
    'Support Messages',
    'User conversations from the support widget.',
    'Refresh',
    'Conversations',
    'Messages',
    'Latest Activity',
    'No support conversations yet.',
    'No open conversations.',
    'No closed conversations.',
    'Open the support widget messages will appear here.',
    'No messages',
    'Not available',
    'Attachment',
    'Reply',
    'Write an answer',
    'Send Reply',
    'Sending...',
    'Open',
    'Closed',
    'Close conversation',
    'Reopen conversation',
    'Updating...',
    'Support team',
    'Customer',
    'Unable to load support conversations',
    'Unable to load support conversation',
    'Unable to send support reply',
    'Unable to update support conversation',
    'Support reply sent.',
    'Support conversation closed.',
    'Support conversation reopened.',
    'Unable to download attachment',
    'Unable to load image preview',
    error,
    message,
    ...conversations.flatMap((conversation) => [conversation.user?.name, conversation.user?.email, lastMessageText(conversation)].filter(Boolean)),
  ]);

  useEffect(() => {
    selectedConversationRef.current = selectedConversation;
  }, [selectedConversation]);

  useEffect(() => {
    activeStatusFilterRef.current = activeStatusFilter;
  }, [activeStatusFilter]);

  const loadConversation = useCallback(
    async (conversationID: number, silent = false) => {
      if (!silent) {
        setIsConversationLoading(true);
      }
      setError('');

      try {
        const loadedConversation = await getAdminSupportConversation(token, conversationID);
        setSelectedConversation(loadedConversation);
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : 'Unable to load support conversation');
        if (loadError instanceof ApiError && loadError.status === 401) {
          onUnauthorized();
        }
      } finally {
        if (!silent) {
          setIsConversationLoading(false);
        }
      }
    },
    [onUnauthorized, token],
  );

  const loadConversations = useCallback(
    async (preferredConversationID?: number | null, silent = false) => {
      if (!silent) {
        setIsLoading(true);
      }
      setError('');

      try {
        const loadedConversations = await listAdminSupportConversations(token);
        setConversations(loadedConversations);
        const statusFilter = activeStatusFilterRef.current;
        const filteredConversations = loadedConversations.filter((conversation) =>
          statusFilter === 'closed' ? conversation.status === 'closed' : conversation.status !== 'closed',
        );
        const nextSelectedID =
          (preferredConversationID && filteredConversations.some((conversation) => conversation.id === preferredConversationID)
            ? preferredConversationID
            : null) ||
          (selectedConversationRef.current?.id &&
          filteredConversations.some((conversation) => conversation.id === selectedConversationRef.current?.id)
            ? selectedConversationRef.current.id
            : null) ||
          filteredConversations.find((conversation) => (conversation.messages?.length || 0) > 0)?.id ||
          filteredConversations[0]?.id ||
          null;

        if (nextSelectedID) {
          const selectedFromList = loadedConversations.find((conversation) => conversation.id === nextSelectedID) || null;
          setSelectedConversation(selectedFromList);
          await loadConversation(nextSelectedID, silent);
        } else {
          setSelectedConversation(null);
        }
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : 'Unable to load support conversations');
        if (loadError instanceof ApiError && loadError.status === 401) {
          onUnauthorized();
        }
      } finally {
        if (!silent) {
          setIsLoading(false);
        }
      }
    },
    [loadConversation, onUnauthorized, token],
  );

  useEffect(() => {
    loadConversations();
  }, [loadConversations]);

  useEffect(() => {
    let isDisposed = false;
    let reconnectTimerID: number | undefined;
    let controller: AbortController | null = null;

    const connect = () => {
      controller = new AbortController();
      streamAdminSupportEvents(token, controller.signal, (event) => {
        const selectedID = selectedConversationRef.current?.id;
        loadConversations(selectedID || event.conversation_id, true);
      }).catch((streamError) => {
        if (isDisposed || isAbortError(streamError)) {
          return;
        }
        if (streamError instanceof ApiError && streamError.status === 401) {
          onUnauthorized();
          return;
        }
        reconnectTimerID = window.setTimeout(connect, reconnectDelayMs);
      });
    };

    connect();

    return () => {
      isDisposed = true;
      controller?.abort();
      if (reconnectTimerID) {
        window.clearTimeout(reconnectTimerID);
      }
    };
  }, [loadConversations, onUnauthorized, token]);

  const handleSelectConversation = (conversationID: number) => {
    const selectedFromList = conversations.find((conversation) => conversation.id === conversationID) || null;
    setSelectedConversation(selectedFromList);
    loadConversation(conversationID);
  };

  const handleStatusFilterChange = (statusFilter: SupportStatusFilter) => {
    activeStatusFilterRef.current = statusFilter;
    setActiveStatusFilter(statusFilter);
    const nextConversation = conversations.find((conversation) =>
      statusFilter === 'closed' ? conversation.status === 'closed' : conversation.status !== 'closed',
    );
    setSelectedConversation(nextConversation || null);
    if (nextConversation) {
      loadConversation(nextConversation.id);
    }
  };

  const handleUpdateStatus = async (status: 'open' | 'closed') => {
    if (!selectedConversationID || isStatusUpdating) {
      return;
    }

    setIsStatusUpdating(true);
    setError('');
    setMessage('');

    try {
      const updatedConversation = await updateAdminSupportConversationStatus(token, selectedConversationID, status);
      const nextFilter: SupportStatusFilter = status === 'closed' ? 'closed' : 'open';
      activeStatusFilterRef.current = nextFilter;
      setActiveStatusFilter(nextFilter);
      setSelectedConversation(updatedConversation);
      setMessage(status === 'closed' ? 'Support conversation closed.' : 'Support conversation reopened.');
      await loadConversations(updatedConversation.id, true);
    } catch (statusError) {
      setError(statusError instanceof Error ? statusError.message : 'Unable to update support conversation');
      if (statusError instanceof ApiError && statusError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsStatusUpdating(false);
    }
  };

  const handleReplySubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selectedConversationID || !replyText.trim() || isReplying) {
      return;
    }

    setIsReplying(true);
    setError('');
    setMessage('');

    try {
      await replyAdminSupportMessage(token, selectedConversationID, replyText.trim());
      setReplyText('');
      setMessage('Support reply sent.');
      await loadConversation(selectedConversationID);
      await loadConversations(selectedConversationID);
    } catch (replyError) {
      setError(replyError instanceof Error ? replyError.message : 'Unable to send support reply');
      if (replyError instanceof ApiError && replyError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsReplying(false);
    }
  };

  const handleDownloadAttachment = async (attachment: SupportAttachment) => {
    setDownloadingAttachmentID(attachment.id);
    setError('');

    try {
      const blob = await downloadSupportAttachment(token, attachment);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = attachment.file_name;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch (downloadError) {
      setError(downloadError instanceof Error ? downloadError.message : 'Unable to download attachment');
      if (downloadError instanceof ApiError && downloadError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setDownloadingAttachmentID(null);
    }
  };

  const selectedIsClosed = selectedConversation?.status === 'closed';

  return (
    <section className="space-y-6">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
          <p className="text-sm text-gray-400">{t('Conversations')}</p>
          <p className="mt-2 text-3xl font-bold text-white">{openConversationCount + closedConversationCount}</p>
        </div>
        <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
          <p className="text-sm text-gray-400">{t('Messages')}</p>
          <p className="mt-2 text-3xl font-bold text-cyan-300">{totalMessages}</p>
        </div>
        <div className="rounded-xl border border-cyan-500/20 bg-white/10 p-5">
          <p className="text-sm text-gray-400">{t('Latest Activity')}</p>
          <p className="mt-2 text-lg font-semibold text-white">{formatDate(conversations[0]?.last_message_at || conversations[0]?.updated_at)}</p>
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

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(18rem,0.75fr)_minmax(0,1.4fr)]">
        <section className="rounded-xl border border-cyan-500/20 bg-white/10 p-4">
          <div className="mb-4 flex items-center justify-between gap-3">
            <div>
              <p className="text-sm font-semibold uppercase tracking-[0.2em] text-cyan-300">{t('Support Messages')}</p>
              <p className="mt-1 text-sm text-gray-400">{t('User conversations from the support widget.')}</p>
            </div>
            <button
              type="button"
              onClick={() => loadConversations(selectedConversation?.id)}
              disabled={isLoading}
              className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-cyan-500/30 text-cyan-100 transition-colors hover:bg-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-60"
              aria-label={t('Refresh')}
            >
              <RefreshCw size={16} className={isLoading ? 'animate-spin' : ''} />
            </button>
          </div>

          <div className="mb-4 grid grid-cols-2 gap-2 rounded-lg border border-cyan-500/15 bg-black/35 p-1">
            <button
              type="button"
              onClick={() => handleStatusFilterChange('open')}
              className={`inline-flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-semibold transition-colors ${
                activeStatusFilter === 'open' ? 'bg-cyan-500 text-black' : 'text-cyan-100 hover:bg-cyan-500/10'
              }`}
            >
              <Unlock size={15} />
              {t('Open')} ({openConversationCount})
            </button>
            <button
              type="button"
              onClick={() => handleStatusFilterChange('closed')}
              className={`inline-flex items-center justify-center gap-2 rounded-md px-3 py-2 text-sm font-semibold transition-colors ${
                activeStatusFilter === 'closed' ? 'bg-cyan-500 text-black' : 'text-cyan-100 hover:bg-cyan-500/10'
              }`}
            >
              <Lock size={15} />
              {t('Closed')} ({closedConversationCount})
            </button>
          </div>

          {isLoading ? (
            <div className="space-y-3" aria-label={t('Conversations')}>
              {[0, 1, 2].map((item) => (
                <div key={item} className="h-20 animate-pulse rounded-xl bg-black/40" />
              ))}
            </div>
          ) : visibleConversations.length === 0 ? (
            <div className="rounded-xl border border-cyan-500/10 bg-black/40 px-4 py-10 text-center">
              <Inbox size={30} className="mx-auto text-cyan-300" />
              <p className="mt-3 text-lg font-semibold text-white">
                {t(conversations.length === 0 ? 'No support conversations yet.' : activeStatusFilter === 'closed' ? 'No closed conversations.' : 'No open conversations.')}
              </p>
              {conversations.length === 0 && <p className="mt-2 text-sm text-gray-400">{t('Open the support widget messages will appear here.')}</p>}
            </div>
          ) : (
            <div className="space-y-2">
              {visibleConversations.map((conversation) => (
                <button
                  key={conversation.id}
                  type="button"
                  onClick={() => handleSelectConversation(conversation.id)}
                  className={`w-full rounded-xl border p-3 text-left transition-colors ${
                    selectedConversation?.id === conversation.id
                      ? 'border-cyan-400/60 bg-cyan-500/15'
                      : 'border-cyan-500/10 bg-black/35 hover:border-cyan-400/35 hover:bg-cyan-500/10'
                  }`}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate font-semibold text-white">{t(customerName(conversation))}</p>
                      <p className="mt-1 truncate text-sm text-gray-400">{conversation.user?.email}</p>
                    </div>
                    <span className="rounded-full bg-cyan-500/15 px-2 py-1 text-xs text-cyan-100">{conversation.messages?.length || 0}</span>
                  </div>
                  <span
                    className={`mt-3 inline-flex rounded-full px-2 py-1 text-xs font-semibold capitalize ${
                      conversation.status === 'closed' ? 'bg-gray-500/15 text-gray-200' : 'bg-emerald-500/15 text-emerald-200'
                    }`}
                  >
                    {t(statusLabel(conversation.status))}
                  </span>
                  <p className="mt-3 line-clamp-2 text-sm text-gray-300">{t(lastMessageText(conversation))}</p>
                  <p className="mt-2 text-xs text-gray-500">{formatDate(conversation.last_message_at || conversation.updated_at)}</p>
                </button>
              ))}
            </div>
          )}
        </section>

        <section className="flex min-h-[36rem] flex-col rounded-xl border border-cyan-500/20 bg-white/10">
          <div className="border-b border-cyan-500/15 px-5 py-4">
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div className="min-w-0">
                <h2 className="truncate text-2xl font-semibold text-white">{selectedConversation ? t(customerName(selectedConversation)) : t('Messages')}</h2>
                {selectedConversation?.user?.email && <p className="mt-1 truncate text-sm text-gray-400">{selectedConversation.user.email}</p>}
              </div>
              <div className="flex flex-wrap items-center gap-2">
                {selectedConversation && (
                  <span
                    className={`inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm font-semibold ${
                      selectedIsClosed
                        ? 'border-gray-400/20 bg-gray-500/10 text-gray-100'
                        : 'border-emerald-400/20 bg-emerald-500/10 text-emerald-100'
                    }`}
                  >
                    {selectedIsClosed ? <Lock size={16} /> : <Unlock size={16} />}
                    {t(statusLabel(selectedConversation.status))}
                  </span>
                )}
                <span className="inline-flex items-center gap-2 rounded-lg border border-cyan-500/20 bg-cyan-500/10 px-3 py-2 text-sm text-cyan-100">
                  <MessageSquare size={16} />
                  {(selectedConversation?.messages || []).length}
                </span>
                {selectedConversation && (
                  <button
                    type="button"
                    onClick={() => handleUpdateStatus(selectedIsClosed ? 'open' : 'closed')}
                    disabled={isStatusUpdating}
                    className="inline-flex items-center justify-center gap-2 rounded-lg border border-cyan-500/30 px-3 py-2 text-sm font-semibold text-cyan-100 transition-colors hover:bg-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {isStatusUpdating ? <Loader2 size={16} className="animate-spin" /> : selectedIsClosed ? <Unlock size={16} /> : <Lock size={16} />}
                    {isStatusUpdating ? t('Updating...') : selectedIsClosed ? t('Reopen conversation') : t('Close conversation')}
                  </button>
                )}
              </div>
            </div>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
            {isConversationLoading ? (
              <div className="space-y-3" aria-label={t('Messages')}>
                {[0, 1, 2, 3].map((item) => (
                  <div key={item} className="h-20 animate-pulse rounded-xl bg-black/40" />
                ))}
              </div>
            ) : !selectedConversation ? (
              <div className="flex h-full min-h-80 items-center justify-center rounded-xl border border-cyan-500/10 bg-black/35 text-center">
                <div>
                  <Inbox size={34} className="mx-auto text-cyan-300" />
                  <p className="mt-3 text-lg font-semibold text-white">{t('No support conversations yet.')}</p>
                </div>
              </div>
            ) : (selectedConversation.messages || []).length === 0 ? (
              <div className="rounded-xl border border-cyan-500/10 bg-black/35 px-4 py-10 text-center text-gray-300">{t('No messages')}</div>
            ) : (
              <div className="space-y-4">
                {(selectedConversation.messages || []).map((supportMessage) => {
                  const isAdminMessage = supportMessage.sender_role === 'admin';
                  return (
                    <article key={supportMessage.id} className={`flex ${isAdminMessage ? 'justify-end' : 'justify-start'}`}>
                      <div
                        className={`max-w-[82%] rounded-xl border px-4 py-3 ${
                          isAdminMessage
                            ? 'border-cyan-400/20 bg-cyan-500 text-black'
                            : 'border-cyan-500/10 bg-black/35 text-gray-100'
                        }`}
                      >
                        <div className="mb-2 flex items-center justify-between gap-4 text-xs font-semibold opacity-75">
                          <span>{isAdminMessage ? t('Support team') : t('Customer')}</span>
                          <span>{formatDate(supportMessage.created_at)}</span>
                        </div>
                        {supportMessage.body && <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">{supportMessage.body}</p>}
                        {(supportMessage.attachments || []).length > 0 && (
                          <div className="mt-3 space-y-2">
                            {(supportMessage.attachments || [])
                              .filter(isImageAttachment)
                              .map((attachment) => (
                                <SupportImageAttachmentPreview
                                  key={attachment.id}
                                  attachment={attachment}
                                  token={token}
                                  isOwnMessage={isAdminMessage}
                                  onDownload={handleDownloadAttachment}
                                  onError={setError}
                                  onUnauthorized={onUnauthorized}
                                />
                              ))}
                            {(supportMessage.attachments || [])
                              .filter((attachment) => !isImageAttachment(attachment))
                              .map((attachment) => (
                                <button
                                  key={attachment.id}
                                  type="button"
                                  onClick={() => handleDownloadAttachment(attachment)}
                                  className={`flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-xs transition-colors ${
                                    isAdminMessage ? 'bg-black/10 hover:bg-black/15' : 'bg-cyan-500/10 hover:bg-cyan-500/15'
                                  }`}
                                >
                                  {downloadingAttachmentID === attachment.id ? (
                                    <Loader2 size={14} className="shrink-0 animate-spin" />
                                  ) : (
                                    <Download size={14} className="shrink-0" />
                                  )}
                                  <span className="min-w-0 flex-1 truncate">{attachment.file_name}</span>
                                  <span className="shrink-0 opacity-70">{formatFileSize(attachment.file_size)}</span>
                                </button>
                              ))}
                          </div>
                        )}
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
          </div>

          <form onSubmit={handleReplySubmit} className="border-t border-cyan-500/15 p-4">
            <label className="block space-y-2 text-sm text-gray-300">
              <span className="inline-flex items-center gap-2">
                <Reply size={15} />
                {t('Reply')}
              </span>
              <textarea
                value={replyText}
                onChange={(event) => setReplyText(event.target.value)}
                disabled={!selectedConversationID || isReplying}
                className="min-h-28 w-full resize-y rounded-lg border border-cyan-500/25 bg-black/60 px-3 py-2 text-sm text-white placeholder-gray-500 transition-colors focus:border-cyan-400 focus:outline-none disabled:cursor-not-allowed disabled:opacity-60"
                placeholder={t('Write an answer')}
                maxLength={4000}
              />
            </label>
            <button
              type="submit"
              disabled={!selectedConversationID || !replyText.trim() || isReplying}
              className="mt-3 inline-flex w-full items-center justify-center gap-2 rounded-lg bg-cyan-500 px-4 py-3 text-sm font-semibold text-black transition-colors hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isReplying ? <Loader2 size={17} className="animate-spin" /> : <Send size={17} />}
              {isReplying ? t('Sending...') : t('Send Reply')}
            </button>
          </form>
        </section>
      </div>
    </section>
  );
};

export default AdminSupportDashboard;
