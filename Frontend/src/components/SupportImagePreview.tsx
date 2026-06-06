import { Download, Loader2, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { downloadSupportAttachment } from '../lib/api';
import type { SupportAttachment } from '../types/api';

interface SupportImageAttachmentPreviewProps {
  attachment: SupportAttachment;
  token: string;
  isOwnMessage: boolean;
  onDownload: (attachment: SupportAttachment) => void;
  onError?: (message: string) => void;
  onUnauthorized?: () => void;
}

interface SelectedImagePreviewProps {
  file: File;
  onRemove: () => void;
  removeLabel: string;
}

const imageContentTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif'];

export const isImageAttachment = (attachment: Pick<SupportAttachment, 'content_type'>) =>
  imageContentTypes.includes(attachment.content_type.toLowerCase().trim());

export const isImageFile = (file: Pick<File, 'type'>) => imageContentTypes.includes(file.type.toLowerCase().trim());

const formatFileSize = (size: number) => {
  if (size < 1024 * 1024) {
    return `${Math.max(1, Math.round(size / 1024))} KB`;
  }

  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
};

export const SupportImageAttachmentPreview = ({
  attachment,
  token,
  isOwnMessage,
  onDownload,
  onError,
  onUnauthorized,
}: SupportImageAttachmentPreviewProps) => {
  const [previewUrl, setPreviewUrl] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const safeAttachmentFileName = sanitizeFileName(attachment.file_name) || 'Attachment';

  useEffect(() => {
    let isMounted = true;
    let nextUrl = '';

    setIsLoading(true);
    setPreviewUrl('');

    downloadSupportAttachment(token, attachment)
      .then((blob) => {
        if (!isMounted) {
          return;
        }
        nextUrl = URL.createObjectURL(blob);
        setPreviewUrl(nextUrl);
      })
      .catch((error) => {
        if (!isMounted) {
          return;
        }
        const status = typeof error === 'object' && error && 'status' in error ? Number(error.status) : 0;
        if (status === 401) {
          onUnauthorized?.();
          return;
        }
        onError?.(error instanceof Error ? error.message : 'Unable to load image preview');
      })
      .finally(() => {
        if (isMounted) {
          setIsLoading(false);
        }
      });

    return () => {
      isMounted = false;
      if (nextUrl) {
        URL.revokeObjectURL(nextUrl);
      }
    };
  }, [attachment, onError, onUnauthorized, token]);

  useEffect(() => {
    if (!isViewerOpen) {
      return undefined;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsViewerOpen(false);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isViewerOpen]);

  if (isLoading) {
    return (
      <div
        className={`flex h-40 w-56 max-w-full items-center justify-center rounded-lg ${
          isOwnMessage ? 'bg-black/10' : 'bg-black/30'
        }`}
      >
        <Loader2 size={20} className="animate-spin opacity-70" />
      </div>
    );
  }

  if (!previewUrl) {
    return (
      <button
        type="button"
        onClick={() => onDownload(attachment)}
        className={`flex w-full items-center gap-2 rounded-lg px-2 py-1 text-left text-xs transition-colors ${
          isOwnMessage ? 'bg-black/10 hover:bg-black/15' : 'bg-black/25 hover:bg-black/35'
        }`}
      >
        <Download size={14} className="shrink-0" />
        <span className="min-w-0 flex-1 truncate">{attachment.file_name}</span>
        <span className="shrink-0 opacity-70">{formatFileSize(attachment.file_size)}</span>
      </button>
    );
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setIsViewerOpen(true)}
        className="group relative block max-w-full overflow-hidden rounded-lg text-left shadow-sm"
      >
        <img src={previewUrl} alt={safeAttachmentFileName} className="h-40 w-56 max-w-full object-cover" />
        <span className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent px-2 pb-2 pt-8 text-xs font-semibold text-white opacity-0 transition-opacity group-hover:opacity-100">
          {safeAttachmentFileName}
        </span>
      </button>

      {isViewerOpen && (
        <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/85 p-4" role="dialog" aria-modal="true">
          <div className="absolute right-4 top-4 flex items-center gap-2">
            <button
              type="button"
              onClick={() => onDownload(attachment)}
              className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-white/10 text-white transition-colors hover:bg-white/20"
              aria-label="Download"
            >
              <Download size={18} />
            </button>
            <button
              type="button"
              onClick={() => setIsViewerOpen(false)}
              className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-white/10 text-white transition-colors hover:bg-white/20"
              aria-label="Close"
            >
              <X size={20} />
            </button>
          </div>
          <img src={previewUrl} alt={safeAttachmentFileName} className="max-h-[88vh] max-w-[92vw] rounded-lg object-contain shadow-2xl" />
        </div>
      )}
    </>
  );
};

const sanitizeFileName = (name: string) => name.replace(/[\r\n<>"'`]/g, '').trim();

export const SelectedImagePreview = ({ file, onRemove, removeLabel }: SelectedImagePreviewProps) => {
  const [previewUrl, setPreviewUrl] = useState('');

  const safeFileName = sanitizeFileName(file.name) || 'Selected image';

  useEffect(() => {
    if (!isImageFile(file)) {
      setPreviewUrl('');
      return undefined;
    }

    const objectUrl = URL.createObjectURL(file);
    setPreviewUrl(objectUrl);

    return () => {
      URL.revokeObjectURL(objectUrl);
    };
  }, [file]);

  if (!previewUrl.startsWith('blob:')) {
    return null;
  }

  return (
    <div className="group relative h-20 w-20 overflow-hidden rounded-lg border border-cyan-500/20 bg-black/35">
      <img src={previewUrl} alt={safeFileName} className="h-full w-full object-cover" />
      <button
        type="button"
        onClick={onRemove}
        className="absolute right-1 top-1 inline-flex h-6 w-6 items-center justify-center rounded-md bg-black/70 text-white transition-colors hover:bg-black"
        aria-label={removeLabel}
      >
        <X size={13} />
      </button>
      <span className="pointer-events-none absolute inset-x-0 bottom-0 truncate bg-black/65 px-1.5 py-1 text-[10px] font-semibold text-white">
        {safeFileName}
      </span>
    </div>
  );
};