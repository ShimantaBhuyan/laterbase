import { useState, useRef, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useLink, useDeleteLink, useReprocessLink, useUpdateLink } from '../hooks/useLinks';
import { useGroups } from '../hooks/useGroups';
import { useMediaQuery } from '../hooks/useMediaQuery';
import StatusBadge from '../components/StatusBadge';
import ContentPreview from '../components/ContentPreview';
import GroupSelector from '../components/GroupSelector';

function getDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return url;
  }
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '—';
  return new Date(dateStr).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function LinkDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isDesktop = useMediaQuery('(min-width: 768px)');

  const { data: link, isLoading } = useLink(id!);
  const { data: groups } = useGroups();
  const deleteLink = useDeleteLink();
  const reprocessLink = useReprocessLink();
  const updateLink = useUpdateLink();

  // Notes state
  const [notesDraft, setNotesDraft] = useState('');
  const [isEditingNotes, setIsEditingNotes] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Sync draft when link loads
  useEffect(() => {
    if (link) {
      setNotesDraft(link.notes ?? '');
    }
  }, [link]);

  // Focus textarea when entering edit mode
  useEffect(() => {
    if (isEditingNotes && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [isEditingNotes]);

  if (isLoading) {
    return (
      <div style={{ maxWidth: 720, margin: '0 auto', paddingTop: 20 }}>
        <div style={{ color: 'var(--color-text-muted)', fontSize: 14 }}>Loading…</div>
      </div>
    );
  }

  if (!link) {
    return (
      <div style={{ maxWidth: 720, margin: '0 auto', paddingTop: 20 }}>
        <div style={{ color: 'var(--color-red)', fontSize: 14 }}>Link not found</div>
      </div>
    );
  }

  const handleDelete = () => {
    if (window.confirm('Delete this link?')) {
      deleteLink.mutate(link.id, {
        onSuccess: () => navigate('/'),
      });
    }
  };

  const handleCopyUrl = () => {
    navigator.clipboard.writeText(link.url);
  };

  const handleSaveNotes = () => {
    updateLink.mutate(
      { id: link.id, data: { notes: notesDraft.trim() || '' } },
      { onSuccess: () => setIsEditingNotes(false) }
    );
  };

  const handleClearNotes = () => {
    if (window.confirm('Clear your notes?')) {
      setNotesDraft('');
      updateLink.mutate(
        { id: link.id, data: { notes: '' } },
        { onSuccess: () => setIsEditingNotes(false) }
      );
    }
  };

  const handleCancelNotes = () => {
    setNotesDraft(link.notes ?? '');
    setIsEditingNotes(false);
  };

  // ── Notes section ────────────────────────────
  const notesSection = (
    <div
      style={{
        marginBottom: 12,
        padding: '8px 12px',
        borderRadius: 'var(--radius-md)',
        border: '1px solid var(--color-border)',
        background: 'var(--color-surface)',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: isEditingNotes ? 8 : 0,
        }}
      >
        <span style={{ fontSize: 12, fontWeight: 590, color: 'var(--color-text-muted)' }}>
          Notes
        </span>
        {!isEditingNotes && (
          <button
            onClick={() => setIsEditingNotes(true)}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--color-accent)',
              fontSize: 12,
              cursor: 'pointer',
              padding: 0,
            }}
          >
            {link.notes ? 'Edit' : 'Add'}
          </button>
        )}
      </div>

      {isEditingNotes ? (
        <div>
          <textarea
            ref={textareaRef}
            value={notesDraft}
            onChange={(e) => setNotesDraft(e.target.value)}
            placeholder="Write your notes here…"
            rows={3}
            style={{
              width: '100%',
              resize: 'vertical',
              padding: '6px 8px',
              fontSize: 13,
              fontFamily: 'inherit',
              lineHeight: 1.5,
              borderRadius: 'var(--radius-sm)',
              border: '1px solid var(--color-border)',
              background: 'var(--color-canvas)',
              color: 'var(--color-text-primary)',
              outline: 'none',
              boxSizing: 'border-box',
            }}
          />
          <div
            style={{
              display: 'flex',
              gap: 6,
              marginTop: 6,
            }}
          >
            <button
              onClick={handleSaveNotes}
              disabled={updateLink.isPending}
              style={{
                padding: '4px 10px',
                borderRadius: 'var(--radius-sm)',
                border: 'none',
                background: 'var(--color-accent)',
                color: '#fff',
                fontSize: 12,
                fontWeight: 510,
                cursor: updateLink.isPending ? 'not-allowed' : 'pointer',
                opacity: updateLink.isPending ? 0.5 : 1,
              }}
            >
              {updateLink.isPending ? 'Saving…' : 'Save'}
            </button>
            <button
              onClick={handleCancelNotes}
              style={{
                padding: '4px 10px',
                borderRadius: 'var(--radius-sm)',
                border: '1px solid var(--color-border)',
                background: 'transparent',
                color: 'var(--color-text-secondary)',
                fontSize: 12,
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            {notesDraft.trim() && (
              <button
                onClick={handleClearNotes}
                style={{
                  padding: '4px 10px',
                  borderRadius: 'var(--radius-sm)',
                  border: '1px solid var(--color-red)',
                  background: 'transparent',
                  color: 'var(--color-red)',
                  fontSize: 12,
                  cursor: 'pointer',
                  marginLeft: 'auto',
                }}
              >
                Clear
              </button>
            )}
          </div>
        </div>
      ) : link.notes ? (
        <div
          style={{
            fontSize: 13,
            lineHeight: 1.5,
            color: 'var(--color-text-primary)',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}
        >
          {link.notes}
        </div>
      ) : null}
    </div>
  );

  // ── Left panel (details) ────────────────────────
  const leftPanel = (
    <>
      {/* Back arrow */}
      <button
        onClick={() => navigate('/')}
        style={{
          background: 'none',
          border: 'none',
          color: 'var(--color-text-secondary)',
          fontSize: 14,
          cursor: 'pointer',
          padding: '0 0 8px 0',
          display: 'flex',
          alignItems: 'center',
          gap: 6,
        }}
      >
        ← Dashboard
      </button>

      {/* Title */}
      <h1
        style={{
          fontSize: isDesktop ? 20 : 22,
          fontWeight: 590,
          color: 'var(--color-text-primary)',
          margin: '0 0 8px 0',
          lineHeight: 1.3,
          wordBreak: 'break-word',
        }}
      >
        {link.title || link.url}
      </h1>

      {/* Status + Group + Actions row */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          marginBottom: link.summary ? 8 : 12,
          flexWrap: 'wrap',
        }}
      >
        <StatusBadge status={link.status} />
        {groups && (
          <GroupSelector
            linkId={link.id}
            currentGroupIds={link.group_ids}
            groups={groups}
          />
        )}

        {/* Spacer on mobile; stack on desktop */}
        {!isDesktop && <div style={{ flex: 1 }} />}

        {/* Action buttons — compact */}
        <a
          href={link.url}
          target="_blank"
          rel="noopener noreferrer"
          style={{
            fontSize: 12,
            fontWeight: 510,
            color: 'var(--color-accent)',
            textDecoration: 'none',
            padding: '4px 10px',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--color-accent)',
          }}
        >
          Open
        </a>
        <button
          onClick={handleCopyUrl}
          style={{
            padding: '4px 10px',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--color-border)',
            background: 'transparent',
            color: 'var(--color-text-secondary)',
            fontSize: 12,
            fontWeight: 510,
            cursor: 'pointer',
          }}
        >
          Copy URL
        </button>
        <button
          onClick={() => reprocessLink.mutate(link.id)}
          disabled={reprocessLink.isPending}
          style={{
            padding: '4px 10px',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--color-amber)',
            background: 'transparent',
            color: 'var(--color-amber)',
            fontSize: 12,
            fontWeight: 510,
            cursor: reprocessLink.isPending ? 'not-allowed' : 'pointer',
            opacity: reprocessLink.isPending ? 0.5 : 1,
          }}
        >
          {reprocessLink.isPending ? '…' : 'Reprocess'}
        </button>
        <button
          onClick={handleDelete}
          style={{
            padding: '4px 10px',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--color-red)',
            background: 'transparent',
            color: 'var(--color-red)',
            fontSize: 12,
            fontWeight: 510,
            cursor: 'pointer',
          }}
        >
          Delete
        </button>
      </div>

      {/* Summary */}
      {link.summary && (
        <div
          style={{
            fontSize: 14,
            color: 'var(--color-text-secondary)',
            lineHeight: 1.5,
            marginBottom: 8,
            padding: '8px 12px',
            background: 'var(--color-surface)',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--color-border)',
          }}
        >
          {link.summary}
        </div>
      )}

      {/* Notes */}
      {notesSection}

      {/* Compact inline metadata */}
      <div
        style={{
          display: 'flex',
          fontSize: 12,
          color: 'var(--color-text-muted)',
          marginBottom: 12,
          flexWrap: 'wrap',
          flexDirection: isDesktop ? 'column' : 'row',
          gap: isDesktop ? 4 : 16,
        }}
      >
        <span>
          Domain:{' '}
          <span style={{ color: 'var(--color-text-secondary)' }}>
            {getDomain(link.url)}
          </span>
        </span>
        <span>
          Source:{' '}
          <span style={{ color: 'var(--color-text-secondary)' }}>
            {link.source_type || '—'}
          </span>
        </span>
        <span>
          Groups:{' '}
          <span style={{ color: 'var(--color-text-secondary)' }}>
            {link.group_ids.length > 0
              ? link.group_ids
                  .map((gid) => groups?.find((g) => g.id === gid)?.name)
                  .filter(Boolean)
                  .join(', ')
              : 'None'}
          </span>
        </span>
        <span>
          Created:{' '}
          <span style={{ color: 'var(--color-text-secondary)' }}>
            {formatDate(link.created_at)}
          </span>
        </span>
        {link.processed_at && (
          <span>
            Processed:{' '}
            <span style={{ color: 'var(--color-text-secondary)' }}>
              {formatDate(link.processed_at)}
            </span>
          </span>
        )}
      </div>

      {/* On mobile, content preview goes here (below details) */}
      {!isDesktop && link.status === 'ready' && (
        <div style={{ marginBottom: 16, height: '50vh' }}>
          <ContentPreview url={link.url} contentText={link.content_text} />
        </div>
      )}
    </>
  );

  // ── Right panel (content preview) ──────────────
  const rightPanel =
    isDesktop && link.status === 'ready' ? (
      <div
        style={{
          flex: 1,
          height: '100%',
          overflow: 'hidden',
        }}
      >
        <ContentPreview url={link.url} contentText={link.content_text} />
      </div>
    ) : null;

  // ── Layout ───────────────────────────────────────
  if (isDesktop && link.status === 'ready') {
    return (
      <div
        style={{
          display: 'flex',
          gap: 24,
          height: 'calc(100vh - 88px)', // 48px header + 20px padding top + 20px padding bottom
          overflow: 'hidden',
        }}
      >
        {/* Left panel — fixed width, scrollable */}
        <div
          style={{
            width: 320,
            flexShrink: 0,
            overflow: 'auto',
            paddingRight: 8,
          }}
        >
          {leftPanel}
        </div>

        {/* Right panel — fills remaining space */}
        {rightPanel}
      </div>
    );
  }

  // Mobile / non-ready: single column
  return (
    <div style={{ maxWidth: 720, margin: '0 auto' }}>
      {leftPanel}
    </div>
  );
}
