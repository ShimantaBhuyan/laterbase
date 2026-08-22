import { useState, useEffect, useRef } from 'react';
import axios from 'axios';
import { useCreateLink } from '../hooks/useLinks';

interface AddLinkDialogProps {
  onClose: () => void;
}

export default function AddLinkDialog({ onClose }: AddLinkDialogProps) {
  const [url, setUrl] = useState('');
  const [error, setError] = useState<string | null>(null);
  const createLink = useCreateLink();
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();

    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [onClose]);

  const handleSave = async () => {
    if (!url.trim()) return;
    setError(null);
    try {
      await createLink.mutateAsync(url.trim());
      onClose();
    } catch (err: unknown) {
      let msg = 'Failed to add link. Please check the URL and try again.';
      if (axios.isAxiosError(err)) {
        const serverMsg = (err.response?.data as { error?: string } | undefined)?.error;
        if (serverMsg) msg = serverMsg;
      }
      setError(msg);
    }
  };

  return (
    <div
      role="presentation"
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 'var(--z-modal-backdrop)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className="card-surface"
        role="dialog"
        aria-modal="true"
        aria-labelledby="add-link-title"
        style={{
          width: 480,
          maxWidth: '90vw',
          padding: 24,
          background: 'var(--color-surface-elevated)',
        }}
      >
        <h2
          id="add-link-title"
          style={{
            fontSize: 16,
            fontWeight: 590,
            color: 'var(--color-text-secondary)',
            margin: '0 0 16px 0',
          }}
        >
          Add Link
        </h2>

        <label htmlFor="add-link-url" style={{ display: 'none' }}>URL</label>
        <input
          ref={inputRef}
          id="add-link-url"
          className="ghost"
          type="url"
          placeholder="Paste a link..."
          value={url}
          onChange={(e) => {
            setUrl(e.target.value);
            setError(null);
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSave();
          }}
          style={{
            width: '100%',
            padding: '10px 12px',
            fontSize: 14,
            marginBottom: 12,
            boxSizing: 'border-box',
          }}
        />

        {error && (
          <div
            style={{
              fontSize: 12,
              color: 'var(--color-red)',
              marginBottom: 12,
            }}
          >
            {error}
          </div>
        )}

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
          <button
            onClick={onClose}
            style={{
              padding: '8px 16px',
              borderRadius: 'var(--radius-md)',
              border: 'none',
              background: 'var(--color-surface)',
              color: 'var(--color-text-secondary)',
              fontSize: 13,
              fontWeight: 510,
              cursor: 'pointer',
            }}
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={createLink.isPending || !url.trim()}
            style={{
              padding: '8px 16px',
              borderRadius: 'var(--radius-md)',
              border: 'none',
              background: createLink.isPending || !url.trim()
                ? 'var(--color-accent)'
                : 'var(--color-accent)',
              color: '#fff',
              fontSize: 13,
              fontWeight: 510,
              cursor: createLink.isPending ? 'not-allowed' : 'pointer',
              opacity: createLink.isPending || !url.trim() ? 0.5 : 1,
            }}
          >
            {createLink.isPending ? 'Saving…' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  );
}
