import { useRef, useState, useEffect } from 'react';

interface ContentPreviewProps {
  url: string;
  contentText?: string | null;
}

export default function ContentPreview({ url, contentText }: ContentPreviewProps) {
  const [showContentText, setShowContentText] = useState(false);
  const [iframeLoadState, setIframeLoadState] = useState<'loading' | 'loaded'>('loading');

  // Reset state when URL changes
  useEffect(() => {
    setShowContentText(false);
    setIframeLoadState('loading');
  }, [url]);

  if (!url) {
    return (
      <div
        style={{
          padding: 16,
          color: 'var(--color-text-muted)',
          fontSize: 13,
          textAlign: 'center',
        }}
      >
        No URL available
      </div>
    );
  }

  return (
    <div
      style={{
        width: '100%',
        height: '100%',
        borderRadius: 'var(--radius-lg)',
        overflow: 'hidden',
        border: '1px solid var(--color-border)',
        background: 'var(--color-surface)',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Toolbar */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 8,
          padding: '6px 12px',
          borderBottom: '1px solid var(--color-border)',
          background: 'var(--color-canvas)',
          flexShrink: 0,
        }}
      >
        <span
          style={{
            fontSize: 12,
            color: 'var(--color-text-muted)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            flex: 1,
          }}
        >
          {url}
        </span>

        {contentText && (
          <button
            onClick={() => setShowContentText(!showContentText)}
            style={{
              fontSize: 11,
              color: 'var(--color-accent)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              padding: '2px 6px',
              whiteSpace: 'nowrap',
            }}
          >
            {showContentText ? 'Hide text' : 'Show text'}
          </button>
        )}

        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          style={{
            fontSize: 12,
            color: 'var(--color-accent)',
            textDecoration: 'none',
            fontWeight: 510,
            whiteSpace: 'nowrap',
          }}
        >
          Open in browser ↗
        </a>
      </div>

      {/* Content area */}
      {showContentText && contentText ? (
        <div
          style={{
            flex: 1,
            overflow: 'auto',
            padding: 16,
            fontSize: 13,
            lineHeight: 1.6,
            color: 'var(--color-text-primary)',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontFamily: 'ui-monospace, "SF Mono", "Fira Code", Menlo, Consolas, monospace',
          }}
        >
          {contentText}
        </div>
      ) : (
        <div
          style={{
            flex: 1,
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          <iframe
            src={url}
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
            title="Link preview"
            onLoad={() => setIframeLoadState('loaded')}
            style={{
              width: '100%',
              height: '100%',
              border: 'none',
            }}
          />

          {iframeLoadState === 'loading' && (
            <div
              style={{
                position: 'absolute',
                inset: 0,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: 'var(--color-surface)',
                color: 'var(--color-text-muted)',
                fontSize: 13,
                pointerEvents: 'none',
              }}
            >
              Loading…
            </div>
          )}
        </div>
      )}
    </div>
  );
}
