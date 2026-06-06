import { Download, Loader2, MessageCircle, Paperclip, RefreshCw, Send, X } from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState, type ChangeEvent, type FormEvent } from 'react';
import { useTranslation } from '../i18n/TranslationContext';
import { ApiError, downloadSupportAttachment, getSupportConversation, sendSupportMessage, streamSupportEvents } from '../lib/api';
import type { SupportAttachment, SupportConversation, User } from '../types/api';
import { isImageAttachment, isImageFile, SelectedImagePreview, SupportImageAttachmentPreview } from './SupportImagePreview';

interface SupportChatWidgetProps {
  token?: string;
  user: User | null;
  onUnauthorized: () => void;
}

const acceptedSupportFileTypes = [
  'image/jpeg',
  'image/png',
  'image/webp',
  'image/gif',
  'application/pdf',
  'text/plain',
  'text/csv',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'application/vnd.ms-powerpoint',
  'application/vnd.openxmlformats-officedocument.presentationml.presentation',
];

const maxSelectedFiles = 5;
const reconnectDelayMs = 3000;

const supportFileAccept = acceptedSupportFileTypes.join(',');

const dateTimeFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
});

const formatDateTime = (value?: string) => {
  if (!value) {
    return '';
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? '' : dateTimeFormatter.format(date);
};

const formatFileSize = (size: number) => {
  if (size < 1024 * 1024) {
    return `${Math.max(1, Math.round(size / 1024))} KB`;
  }

  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
};

const sanitizeFileName = (name: string) => name.replace(/[\r\n<>"'`]/g, '').trim();

const getDisplayFileName = (name: string, fallback = 'File') => sanitizeFileName(name) || fallback;

const isAcceptedSupportFile = (file: File) => acceptedSupportFileTypes.includes(file.type);

const isAbortError = (error: unknown) => error instanceof DOMException && error.name === 'AbortError';

const SupportChatWidget = ({ token, user, onUnauthorized }: SupportChatWidgetProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [conversation, setConversation] = useState<SupportConversation | null>(null);
  const [messageText, setMessageText] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [isAdminOnline, setIsAdminOnline] = useState(false);
  const [error, setError] = useState('');
  const [downloadingAttachmentID, setDownloadingAttachmentID] = useState<number | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messages = useMemo(() => conversation?.messages || [], [conversation?.messages]);
  const { t } = useTranslation([
    'Support',
    'AutoRent support',
    'Online',
    'Offline',
    'Open support chat',
    'Close support chat',
    'Sign in to contact support.',
    'No messages yet.',
    'Write a message',
    'Attach files',
    'Remove file',
    'Send',
    'Sending...',
    'Refresh',
    'You',
    'Support team',
    'Unable to load support conversation',
    'Unable to send support message',
    'Unable to download attachment',
    'Unable to load image preview',
    'Only images, PDF, text, and Office files can be uploaded.',
    'You can attach up to 5 files.',
    error,
  ]);

  const loadConversation = useCallback(
    async (silent = false) => {
      if (!token) {
        setConversation(null);
        setIsAdminOnline(false);
        return;
      }

      if (!silent) {
        setIsLoading(true);
      }
      setError('');

      try {
        const loadedConversation = await getSupportConversation(token);
        setConversation(loadedConversation);
      } catch (loadError) {
        setError(loadError instanceof Error ? loadError.message : 'Unable to load support conversation');
        if (loadError instanceof ApiError && loadError.status === 401) {
          onUnauthorized();
        }
      } finally {
        if (!silent) {
          setIsLoading(false);
        }
      }
    },
    [onUnauthorized, token],
  );

  useEffect(() => {
    if (isOpen) {
      loadConversation();
    }
  }, [isOpen, loadConversation]);

  useEffect(() => {
    if (!isOpen || !token) {
      return undefined;
    }

    let isDisposed = false;
    let reconnectTimerID: number | undefined;
    let controller: AbortController | null = null;

    const connect = () => {
      controller = new AbortController();
      streamSupportEvents(token, controller.signal, (event) => {
        if (event.event_type === 'presence') {
          setIsAdminOnline(Boolean(event.admin_online));
          return;
        }
        loadConversation(true);
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
      setIsAdminOnline(false);
      if (reconnectTimerID) {
        window.clearTimeout(reconnectTimerID);
      }
    };
  }, [isOpen, loadConversation, onUnauthorized, token]);

  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
    }
  }, [isOpen, messages.length]);

  const handleFileChange = (event: ChangeEvent<HTMLInputElement>) => {
    const input = event.currentTarget;
    const files = Array.from(input.files ?? []);
    input.value = '';

    if (files.some((file) => !isAcceptedSupportFile(file))) {
      setError('Only images, PDF, text, and Office files can be uploaded.');
      return;
    }

    setSelectedFiles((current) => {
      const nextFiles = [...current, ...files].slice(0, maxSelectedFiles);
      if (current.length + files.length > maxSelectedFiles) {
        setError('You can attach up to 5 files.');
      } else {
        setError('');
      }
      return nextFiles;
    });
  };

  const removeFile = (index: number) => {
    setSelectedFiles((current) => current.filter((_, fileIndex) => fileIndex !== index));
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!token || isSending) {
      return;
    }

    const trimmedMessage = messageText.trim();
    if (!trimmedMessage && selectedFiles.length === 0) {
      return;
    }

    setIsSending(true);
    setError('');

    try {
      await sendSupportMessage(token, trimmedMessage, selectedFiles);
      setMessageText('');
      setSelectedFiles([]);
      await loadConversation(true);
    } catch (sendError) {
      setError(sendError instanceof Error ? sendError.message : 'Unable to send support message');
      if (sendError instanceof ApiError && sendError.status === 401) {
        onUnauthorized();
      }
    } finally {
      setIsSending(false);
    }
  };

  const handleDownloadAttachment = async (attachment: SupportAttachment) => {
    if (!token) {
      return;
    }

    setDownloadingAttachmentID(attachment.id);
    setError('');

    try {
      const blob = await downloadSupportAttachment(token, attachment);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = getDisplayFileName(attachment.file_name, 'attachment');
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

  return (
    <div className="fixed bottom-5 right-4 z-50 sm:bottom-6 sm:right-6">
      {isOpen && (
        <section className="mb-4 flex max-h-[min(42rem,calc(100vh-7rem))] w-[calc(100vw-2rem)] flex-col overflow-hidden rounded-xl border border-violet-300/25 bg-gray-950 shadow-2xl shadow-black/45 sm:w-96">
          <header className="flex items-center justify-between gap-3 bg-violet-600 px-4 py-3 text-white">
            <div className="flex min-w-0 items-center gap-3">
              <span className="relative inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white text-violet-700">
                <MessageCircle size={21} />
                <span
                  className={`absolute bottom-0 right-0 h-3.5 w-3.5 rounded-full border-2 border-violet-600 ${
                    isAdminOnline ? 'bg-emerald-400' : 'bg-gray-400'
                  }`}
                  aria-hidden="true"
                />
              </span>
              <div className="min-w-0">
                <h2 className="truncate text-base font-bold">{t('AutoRent support')}</h2>
                <p className="truncate text-xs text-violet-100">{t(isAdminOnline ? 'Online' : 'Offline')}</p>
              </div>
            </div>
            <div className="flex items-center gap-1">
              {token && (
                <button
                  type="button"
                  onClick={() => loadConversation()}
                  disabled={isLoading}
                  className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-violet-50 transition-colors hover:bg-white/10 disabled:opacity-60"
                  aria-label={t('Refresh')}
                >
                  <RefreshCw size={17} className={isLoading ? 'animate-spin' : ''} />
                </button>
              )}
              <button
                type="button"
                onClick={() => setIsOpen(false)}
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg text-violet-50 transition-colors hover:bg-white/10"
                aria-label={t('Close support chat')}
              >
                <X size={18} />
              </button>
            </div>
          </header>

          <div className="min-h-0 flex-1 overflow-y-auto bg-gray-950 px-4 py-4">
            {!token || !user ? (
              <div className="rounded-lg border border-cyan-500/15 bg-cyan-500/10 px-4 py-5 text-sm text-cyan-50">
                {t('Sign in to contact support.')}
              </div>
            ) : isLoading ? (
              <div className="space-y-3" aria-label={t('Support')}>
                {[0, 1, 2].map((item) => (
                  <div key={item} className="h-16 animate-pulse rounded-lg bg-white/10" />
                ))}
              </div>
            ) : messages.length === 0 ? (
              <div className="rounded-lg border border-white/10 bg-white/5 px-4 py-6 text-center text-sm text-gray-300">
                {t('No messages yet.')}
              </div>
            ) : (
              <div className="space-y-3">
                {messages.map((message) => {
                  const isUserMessage = message.sender_role === 'user';
                  return (
                    <article key={message.id} className={`flex ${isUserMessage ? 'justify-end' : 'justify-start'}`}>
                      <div
                        className={`max-w-[82%] rounded-xl px-3 py-2 text-sm shadow-sm ${
                          isUserMessage ? 'bg-cyan-500 text-black' : 'bg-white/10 text-gray-100'
                        }`}
                      >
                        <div className="mb-1 flex items-center justify-between gap-3 text-[11px] font-semibold opacity-75">
                          <span>{isUserMessage ? t('You') : t('Support team')}</span>
                          <span>{formatDateTime(message.created_at)}</span>
                        </div>
                        {message.body && <p className="whitespace-pre-wrap break-words leading-relaxed">{message.body}</p>}
                        {(message.attachments || []).length > 0 && (
                          <div className="mt-2 space-y-2">
                            {(message.attachments || [])
                              .filter(isImageAttachment)
                              .map((attachment) => (
                                <SupportImageAttachmentPreview
                                  key={attachment.id}
                                  attachment={attachment}
                                  token={token}
                                  isOwnMessage={isUserMessage}
                                  onDownload={handleDownloadAttachment}
                                  onError={setError}
                                  onUnauthorized={onUnauthorized}
                                />
                              ))}
                            {(message.attachments || [])
                              .filter((attachment) => !isImageAttachment(attachment))
                              .map((attachment) => (
                                <button
                                  key={attachment.id}
                                  type="button"
                                  onClick={() => handleDownloadAttachment(attachment)}
                                  className={`flex w-full items-center gap-2 rounded-lg px-2 py-1 text-left text-xs transition-colors ${
                                    isUserMessage ? 'bg-black/10 hover:bg-black/15' : 'bg-black/25 hover:bg-black/35'
                                  }`}
                                >
                                  {downloadingAttachmentID === attachment.id ? (
                                    <Loader2 size={14} className="shrink-0 animate-spin" />
                                  ) : (
                                    <Download size={14} className="shrink-0" />
                                  )}
                                  <span className="min-w-0 flex-1 truncate">{getDisplayFileName(attachment.file_name, t('Attachment'))}</span>
                                  <span className="shrink-0 opacity-70">{formatFileSize(attachment.file_size)}</span>
                                </button>
                              ))}
                          </div>
                        )}
                      </div>
                    </article>
                  );
                })}
                <div ref={messagesEndRef} />
              </div>
            )}
          </div>

          {error && (
            <div className="border-t border-red-400/20 bg-red-500/10 px-4 py-2 text-xs text-red-100" role="alert">
              {t(error)}
            </div>
          )}

          <form onSubmit={handleSubmit} className="border-t border-white/10 bg-gray-900 px-3 py-3">
            {selectedFiles.length > 0 && (
              <div className="mb-2 flex flex-wrap gap-2">
                {selectedFiles.map((file, index) => (
                  isImageFile(file) ? (
                    <SelectedImagePreview
                      key={`${file.name}-${file.lastModified}-${index}`}
                      file={file}
                      onRemove={() => removeFile(index)}
                      removeLabel={t('Remove file')}
                    />
                  ) : (
                    <span
                      key={`${file.name}-${file.lastModified}-${index}`}
                      className="inline-flex max-w-full items-center gap-2 rounded-lg border border-cyan-500/20 bg-cyan-500/10 px-2 py-1 text-xs text-cyan-50"
                    >
                      <span className="max-w-44 truncate">{getDisplayFileName(file.name, t('Selected file'))}</span>
                      <button
                        type="button"
                        onClick={() => removeFile(index)}
                        className="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-md hover:bg-white/10"
                        aria-label={t('Remove file')}
                      >
                        <X size={12} />
                      </button>
                    </span>
                  )
                ))}
              </div>
            )}

            <div className="flex items-end gap-2">
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={!token || isSending}
                className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border border-white/10 text-cyan-100 transition-colors hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-45"
                aria-label={t('Attach files')}
              >
                <Paperclip size={18} />
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept={supportFileAccept}
                multiple
                className="sr-only"
                onChange={handleFileChange}
              />
              <textarea
                value={messageText}
                onChange={(event) => setMessageText(event.target.value)}
                disabled={!token || isSending}
                placeholder={t('Write a message')}
                rows={1}
                maxLength={4000}
                className="min-h-11 flex-1 resize-none rounded-lg border border-white/10 bg-black/35 px-3 py-3 text-sm text-white placeholder-gray-500 outline-none transition-colors focus:border-cyan-400 disabled:cursor-not-allowed disabled:opacity-60"
              />
              <button
                type="submit"
                disabled={!token || isSending || (!messageText.trim() && selectedFiles.length === 0)}
                className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-cyan-500 text-black transition-colors hover:bg-cyan-400 disabled:cursor-not-allowed disabled:opacity-45"
                aria-label={isSending ? t('Sending...') : t('Send')}
              >
                {isSending ? <Loader2 size={18} className="animate-spin" /> : <Send size={18} />}
              </button>
            </div>
          </form>
        </section>
      )}

      <button
        type="button"
        onClick={() => setIsOpen((current) => !current)}
        className="ml-auto flex h-16 w-16 items-center justify-center rounded-full bg-violet-600 text-white shadow-2xl shadow-black/40 transition-transform hover:scale-105 hover:bg-violet-500"
        aria-label={t('Open support chat')}
      >
        {isOpen ? <X size={28} /> : <MessageCircle size={30} />}
      </button>
    </div>
  );
};

export default SupportChatWidget;
