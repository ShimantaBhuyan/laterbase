import { useState, useEffect, useRef, useMemo } from 'react';
import axios from 'axios';
import { useCreateLinksBatch } from '../hooks/useLinks';
import type { BatchCreateResult } from '../api/client';

interface AddLinkDialogProps {
  onClose: () => void;
}

const MAX_URLS = 100;

// Pull every http(s) URL out of free-form pasted text (newlines, spaces,
// commas, markdown, chat exports…). Falls back to bare tokens like
// "example.com/foo" so a plain list of domains still works.
function parseUrls(text: string): string[] {
  const seen = new Set<string>();
  const out: string[] = [];
  const add = (u: string) => {
    const cleaned = u.replace(/[),.;'"\]>]+$/, '');
    if (cleaned && !seen.has(cleaned)) {
      seen.add(cleaned);
      out.push(cleaned);
    }
  };

  const matches = text.match(/https?:\/\/[^\s<>"'`]+/gi);
  if (matches) {
    matches.forEach(add);
  } else {
    text
      .split(/[\s,]+/)
      .filter((t) => /^[\w-]+(\.[\w-]+)+(\/\S*)?$/.test(t))
      .forEach((t) => add(`https://${t}`));
  }
  return out;
}

export default function AddLinkDialog({ onClose }: AddLinkDialogProps) {
  const [text, setText] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [results, setResults] = useState<BatchCreateResult[] | null>(null);
  const createBatch = useCreateLinksBatch();
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const urls = useMemo(() => parseUrls(text), [text]);
  const tooMany = urls.length > MAX_URLS;

  useEffect(() => {
    inputRef.current?.focus();

    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [onClose]);

  const handleSave = async () => {
    if (urls.length === 0 || tooMany || createBatch.isPending) return;
    setError(null);
    setResults(null);
    try {
      const res = await createBatch.mutateAsync(urls);
      const failed = res.results.filter((r) => r.status !== 'created');
      if (failed.length === 0) {
        onClose();
        return;
      }
      setResults(res.results);
      // Leave only the URLs that hit a real error, so they can be retried.
      setText(
        res.results
          .filter((r) => r.status === 'error')
          .map((r) => r.url)
          .join('\n'),
      );
    } catch (err: unknown) {
      let msg = 'Failed to add links. Please try again.';
      if (axios.isAxiosError(err)) {
        const serverMsg = (err.response?.data as { error?: string } | undefined)?.error;
        if (serverMsg) msg = serverMsg;
      }
      setError(msg);
    }
  };

  const created = results?.filter((r) => r.status === 'created').length ?? 0;
  const duplicates = results?.filter((r) => r.status === 'duplicate') ?? [];
  const errors = results?.filter((r) => r.status === 'error') ?? [];

  const label = createBatch.isPending
    ? 'Saving…'
    : urls.length > 1
      ? `Add ${urls.length} links`
      : 'Save';
  const disabled = createBatch.isPending || urls.length === 0 || tooMany;

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
          width: 520,
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
          Add Links
        </h2>

        <label htmlFor="add-link-url" style={{ display: 'none' }}>URLs</label>
        <textarea
          ref={inputRef}
          id="add-link-url"
          className="ghost"
          placeholder={'Paste one or more links…\nhttps://example.com/article\nhttps://github.com/owner/repo'}
          value={text}
          rows={6}
          onChange={(e) => {
            setText(e.target.value);
            setError(null);
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
              e.preventDefault();
              handleSave();
            } else if (e.key === 'Enter' && !e.shiftKey && urls.length <= 1 && !text.includes('\n')) {
              // Single-link fast path: plain Enter still submits.
              e.preventDefault();
              handleSave();
            }
          }}
          style={{
            width: '100%',
            padding: '10px 12px',
            fontSize: 14,
            lineHeight: 1.5,
            marginBottom: 8,
            boxSizing: 'border-box',
            resize: 'vertical',
            minHeight: 120,
            fontFamily: 'inherit',
          }}
        />

        <div
          style={{
            fontSize: 12,
            color: tooMany ? 'var(--color-red)' : 'var(--color-text-muted)',
            marginBottom: 12,
          }}
        >
          {urls.length === 0
            ? 'One link per line, or paste any text containing links. ⌘↵ to save.'
            : tooMany
              ? `${urls.length} links detected — max ${MAX_URLS} at a time.`
              : `${urls.length} link${urls.length === 1 ? '' : 's'} detected · ⌘↵ to save`}
        </div>

        {results && (
          <div
            style={{
              fontSize: 12,
              color: 'var(--color-text-secondary)',
              marginBottom: 12,
              display: 'flex',
              flexDirection: 'column',
              gap: 4,
            }}
          >
            <div>
              Added {created}
              {duplicates.length > 0 && ` · ${duplicates.length} already saved`}
              {errors.length > 0 && ` · ${errors.length} failed`}
            </div>
            {duplicates.length > 0 && (
              <div style={{ color: 'var(--color-text-muted)', wordBreak: 'break-all' }}>
                Skipped: {duplicates.map((d) => d.url).join(', ')}
              </div>
            )}
            {errors.length > 0 && (
              <div style={{ color: 'var(--color-red)' }}>
                Failed links are left above to retry.
              </div>
            )}
          </div>
        )}

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
            {results ? 'Done' : 'Cancel'}
          </button>
          <button
            onClick={handleSave}
            disabled={disabled}
            style={{
              padding: '8px 16px',
              borderRadius: 'var(--radius-md)',
              border: 'none',
              background: 'var(--color-accent)',
              color: '#fff',
              fontSize: 13,
              fontWeight: 510,
              cursor: createBatch.isPending ? 'not-allowed' : 'pointer',
              opacity: disabled ? 0.5 : 1,
            }}
          >
            {label}
          </button>
        </div>
      </div>
    </div>
  );
}
