import { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { askQuestion } from '../api/ask';
import type { AskCitation, AskDonePayload, AskStatus, GroupWithCount } from '../types';
import { useGroups } from '../hooks/useGroups';
import { useTheme } from '../hooks/useTheme';
import { useMediaQuery } from '../hooks/useMediaQuery';

interface AskOverlayProps {
  onClose: () => void;
}

type Phase = 'idle' | 'loading' | 'streaming' | 'done' | 'error';

const EXAMPLE_QUESTIONS = [
  'What did I save about Postgres?',
  'Which GitHub repos have I saved recently?',
  'Summarize what I\u2019ve saved about AI agents',
];

function getDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return url;
  }
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m`;
  if (hours < 24) return `${hours}h`;
  if (days === 1) return 'yesterday';
  if (days < 7) return `${days}d`;
  return new Date(dateStr).toLocaleDateString();
}

export default function AskOverlay({ onClose }: AskOverlayProps) {
  const [question, setQuestion] = useState('');
  const [phase, setPhase] = useState<Phase>('idle');
  const [answer, setAnswer] = useState('');
  const [citations, setCitations] = useState<AskCitation[]>([]);
  const [finalStatus, setFinalStatus] = useState<AskStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [includeFullText, setIncludeFullText] = useState(false);
  const [flashId, setFlashId] = useState<string | null>(null);

  const abortRef = useRef<AbortController | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const rowRefs = useRef<Map<number, HTMLDivElement>>(new Map());
  const flashTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const lastQuestionRef = useRef('');

  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: groups } = useGroups();
  const { theme } = useTheme();
  const isMobile = useMediaQuery('(max-width: 640px)');

  useEffect(() => {
    inputRef.current?.focus();
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = '';
      abortRef.current?.abort();
      if (flashTimerRef.current) clearTimeout(flashTimerRef.current);
    };
  }, []);

  const close = useCallback(() => {
    abortRef.current?.abort();
    onClose();
  }, [onClose]);

  const focusRow = useCallback((n: number) => {
    const row = rowRefs.current.get(n);
    if (!row) return;
    row.scrollIntoView({ behavior: 'smooth', block: 'center' });
    const citation = citations[n - 1];
    if (citation) {
      setFlashId(citation.link_id);
      if (flashTimerRef.current) clearTimeout(flashTimerRef.current);
      flashTimerRef.current = setTimeout(() => setFlashId(null), 1200);
    }
  }, [citations]);

  const submit = useCallback((q: string) => {
    const trimmed = q.trim();
    if (!trimmed || phase === 'loading' || phase === 'streaming') return;

    lastQuestionRef.current = trimmed;
    setError(null);
    setAnswer('');
    setCitations([]);
    setFinalStatus(null);
    setPhase('loading');

    const controller = new AbortController();
    abortRef.current?.abort();
    abortRef.current = controller;

    askQuestion(trimmed, includeFullText, {
      onCitations: (c) => setCitations(c),
      onDelta: (text) => {
        setPhase((p) => (p === 'done' ? p : 'streaming'));
        setAnswer((prev) => prev + text);
      },
      onDone: (payload: AskDonePayload) => {
        setFinalStatus(payload.status);
        if (payload.status === 'ok') setAnswer((prev) => prev || payload.answer);
        setPhase('done');
        queryClient.invalidateQueries({ queryKey: ['stats'] });
      },
      onError: (msg) => {
        setError(msg);
        setPhase('error');
      },
    }, controller.signal).then(() => {
      // Stream ended without an explicit done/error event — treat as complete
      setPhase((p) => (p === 'loading' || p === 'streaming' ? 'done' : p));
      setFinalStatus((s) => s ?? 'ok');
    }).catch(() => {
      // handled via onError / abort
    });
  }, [phase, includeFullText, queryClient]);

  const stop = () => {
    abortRef.current?.abort();
    setPhase(answer ? 'done' : 'idle');
    setFinalStatus((s) => s ?? (answer ? 'ok' : null));
  };

  const openLink = (linkId: string) => {
    abortRef.current?.abort();
    onClose();
    navigate(`/links/${linkId}`);
  };

  const busy = phase === 'loading' || phase === 'streaming';

  // Split answer text so [n] citation markers become clickable
  const answerParts = answer.split(/(\[\d+\])/g);

  return (
    <div
      role="presentation"
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 'var(--z-modal-backdrop)' as unknown as number,
        display: 'flex',
        alignItems: isMobile ? 'stretch' : 'center',
        justifyContent: 'center',
        background: 'rgba(0,0,0,0.6)',
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) close();
      }}
    >
      <div
        className="ask-panel"
        role="dialog"
        aria-modal="true"
        aria-label="Ask your archive"
        style={{
          width: isMobile ? '100%' : 'min(80vw, 1100px)',
          height: isMobile ? '100dvh' : '80vh',
          display: 'flex',
          flexDirection: 'column',
          background: 'var(--color-surface-elevated)',
          border: '1px solid var(--color-border)',
          borderRadius: isMobile ? 0 : 'var(--radius-xl)',
          boxShadow: theme === 'dark' ? 'none' : '0 8px 24px rgba(0,0,0,0.10)',
          overflow: 'hidden',
        }}
        onKeyDown={(e) => {
          if (e.key === 'Escape') close();
        }}
      >
        {/* Prompt row */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            padding: isMobile ? '12px 16px' : '16px 20px',
            borderBottom: '1px solid var(--color-border)',
            flexShrink: 0,
          }}
        >
          <input
            ref={inputRef}
            className="ghost ask-input"
            type="text"
            placeholder="Ask your archive…"
            aria-label="Ask your archive"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') submit(question);
            }}
            disabled={busy}
            style={{
              flex: 1,
              padding: '10px 14px',
              fontSize: 16,
              boxSizing: 'border-box',
              minWidth: 0,
            }}
          />
          <button
            className="ask-fulltext-toggle"
            onClick={() => setIncludeFullText((v) => !v)}
            disabled={busy}
            title="Also search the full page text of matched links (slower)"
            aria-pressed={includeFullText}
            style={{
              padding: '6px 12px',
              borderRadius: 'var(--radius-pill)',
              border: '1px solid',
              borderColor: includeFullText ? 'var(--color-accent)' : 'var(--color-border)',
              background: includeFullText ? 'var(--color-accent-soft, var(--color-surface))' : 'transparent',
              color: includeFullText ? 'var(--color-accent)' : 'var(--color-text-secondary)',
              fontSize: 12,
              fontWeight: 510,
              cursor: busy ? 'not-allowed' : 'pointer',
              whiteSpace: 'nowrap',
              flexShrink: 0,
            }}
          >
            Full text
          </button>
          <button
            onClick={() => (busy ? stop() : submit(question))}
            disabled={!busy && !question.trim()}
            style={{
              padding: '8px 18px',
              borderRadius: 'var(--radius-md)',
              border: 'none',
              background: busy ? 'var(--color-surface-hover)' : 'var(--color-accent)',
              color: busy ? 'var(--color-text-primary)' : '#fff',
              fontSize: 13,
              fontWeight: 510,
              cursor: !busy && !question.trim() ? 'default' : 'pointer',
              opacity: !busy && !question.trim() ? 0.5 : 1,
              whiteSpace: 'nowrap',
              flexShrink: 0,
            }}
          >
            {busy ? 'Stop' : 'Ask'}
          </button>
          <button
            onClick={close}
            aria-label="Close"
            style={{
              width: 30,
              height: 30,
              borderRadius: 'var(--radius-md)',
              border: 'none',
              background: 'transparent',
              color: 'var(--color-text-muted)',
              fontSize: 16,
              cursor: 'pointer',
              lineHeight: 1,
              flexShrink: 0,
            }}
          >
            ✕
          </button>
        </div>

        {/* Scrollable body */}
        <div style={{ flex: 1, overflowY: 'auto', padding: isMobile ? '16px' : '20px 24px' }}>
          {/* Idle: teach with examples */}
          {phase === 'idle' && (
            <div className="ask-fade" style={{ paddingTop: 8 }}>
              <p
                style={{
                  fontSize: 14,
                  color: 'var(--color-text-secondary)',
                  margin: '0 0 14px 0',
                  lineHeight: 1.5,
                  maxWidth: '65ch',
                }}
              >
                Ask a question and laterbase will find the most relevant links in your archive.
              </p>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6, alignItems: 'flex-start' }}>
                {EXAMPLE_QUESTIONS.map((ex) => (
                  <button
                    key={ex}
                    onClick={() => {
                      setQuestion(ex);
                      inputRef.current?.focus();
                    }}
                    style={{
                      padding: '6px 14px',
                      borderRadius: 'var(--radius-pill)',
                      border: '1px solid var(--color-border)',
                      background: 'transparent',
                      color: 'var(--color-text-secondary)',
                      fontSize: 13,
                      fontWeight: 510,
                      cursor: 'pointer',
                      textAlign: 'left',
                    }}
                  >
                    {ex}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Loading skeletons before anything arrives */}
          {phase === 'loading' && citations.length === 0 && (
            <div aria-hidden="true">
              <div style={{ height: 13, width: '70%', background: 'var(--color-border)', borderRadius: 4, marginBottom: 10, animation: 'pulse 1.5s infinite' }} />
              <div style={{ height: 13, width: '45%', background: 'var(--color-border)', borderRadius: 4, marginBottom: 28, animation: 'pulse 1.5s infinite' }} />
              {[0, 1, 2].map((i) => (
                <div key={i} className="card-surface" style={{ padding: 14, marginBottom: 8, opacity: 0.3, animation: 'pulse 1.5s infinite' }}>
                  <div style={{ height: 14, width: '60%', background: 'var(--color-border)', borderRadius: 4, marginBottom: 8 }} />
                  <div style={{ height: 11, width: '40%', background: 'var(--color-border)', borderRadius: 4 }} />
                </div>
              ))}
            </div>
          )}

          {/* Error */}
          {phase === 'error' && (
            <div style={{ paddingTop: 8 }}>
              <p style={{ fontSize: 14, color: 'var(--color-red)', margin: '0 0 12px 0' }}>{error}</p>
              <button
                onClick={() => submit(lastQuestionRef.current)}
                style={{
                  padding: '6px 16px',
                  borderRadius: 'var(--radius-md)',
                  border: '1px solid var(--color-border)',
                  background: 'var(--color-surface)',
                  color: 'var(--color-text-secondary)',
                  fontSize: 12,
                  fontWeight: 510,
                  cursor: 'pointer',
                }}
              >
                Retry
              </button>
            </div>
          )}

          {/* Answer */}
          {(phase === 'streaming' || phase === 'done') && answer && (
            <div className="ask-fade" style={{ marginBottom: 24 }}>
              <p
                style={{
                  fontSize: 14,
                  lineHeight: 1.6,
                  color: 'var(--color-text-primary)',
                  maxWidth: '65ch',
                  margin: 0,
                  whiteSpace: 'pre-wrap',
                }}
              >
                {answerParts.map((part, i) => {
                  const match = part.match(/^\[(\d+)\]$/);
                  if (match) {
                    const n = parseInt(match[1], 10);
                    return (
                      <button
                        key={i}
                        className="ask-marker"
                        onClick={() => focusRow(n)}
                        title={`Jump to result ${n}`}
                        style={{
                          background: 'none',
                          border: 'none',
                          padding: '0 2px',
                          color: n <= citations.length ? 'var(--color-accent)' : 'var(--color-text-muted)',
                          fontSize: 11,
                          fontWeight: 510,
                          verticalAlign: 'super',
                          cursor: n <= citations.length ? 'pointer' : 'default',
                        }}
                      >
                        [{n}]
                      </button>
                    );
                  }
                  return <span key={i}>{part}</span>;
                })}
                {phase === 'streaming' && <span className="ask-cursor" aria-hidden="true">▍</span>}
              </p>
            </div>
          )}

          {/* Honest empty states */}
          {phase === 'done' && finalStatus === 'no_matches' && !answer && (
            <p style={{ fontSize: 14, color: 'var(--color-text-secondary)', margin: '8px 0 0 0' }}>
              Nothing in your archive matches that yet. Save more links and try again.
            </p>
          )}
          {phase === 'done' && finalStatus === 'empty_archive' && (
            <p style={{ fontSize: 14, color: 'var(--color-text-secondary)', margin: '8px 0 0 0' }}>
              Your archive has no processed links yet — add a few and come back.
            </p>
          )}

          {/* Results */}
          {citations.length > 0 && (
            <>
              {citations.length > 0 && (
                <div
                  style={{
                    fontSize: 12,
                    fontWeight: 510,
                    color: 'var(--color-text-muted)',
                    marginBottom: 10,
                    paddingBottom: 8,
                    borderBottom: '1px solid var(--color-border)',
                  }}
                >
                  From your archive
                </div>
              )}
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {citations.map((cit, i) => (
                    <div
                      key={cit.link_id + i}
                      ref={(el) => {
                        if (el) rowRefs.current.set(i + 1, el);
                        else rowRefs.current.delete(i + 1);
                      }}
                      data-flash={flashId === cit.link_id}
                      className="ask-row card-surface"
                      role="button"
                      tabIndex={0}
                      onClick={() => openLink(cit.link_id)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          openLink(cit.link_id);
                        }
                      }}
                      style={{ padding: '12px 14px', display: 'flex', gap: 12, cursor: 'pointer' }}
                    >
                      <span
                        aria-hidden="true"
                        style={{
                          fontSize: 12,
                          fontWeight: 510,
                          color: 'var(--color-accent)',
                          minWidth: 18,
                          paddingTop: 2,
                          flexShrink: 0,
                        }}
                      >
                        {i + 1}
                      </span>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div
                          style={{
                            fontSize: 14,
                            fontWeight: 590,
                            color: 'var(--color-text-primary)',
                            lineHeight: 1.35,
                            marginBottom: 3,
                          }}
                        >
                          {cit.title || cit.url}
                        </div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
                          <span style={{ fontSize: 11, color: 'var(--color-text-muted)' }}>
                            {getDomain(cit.url)} · {timeAgo(cit.created_at)}
                          </span>
                          {linkGroupPills(cit, groups)}
                        </div>
                        {cit.reason && (
                          <div
                            style={{
                              fontSize: 12,
                              fontStyle: 'italic',
                              color: 'var(--color-text-secondary)',
                              marginTop: 4,
                              lineHeight: 1.4,
                            }}
                          >
                            {cit.reason}
                          </div>
                        )}
                      </div>
                    </div>
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

// Renders up to two group pills for a citation row
function linkGroupPills(
  cit: AskCitation,
  groups: GroupWithCount[] | undefined
) {
  if (!groups || cit.group_ids.length === 0) return null;
  return (
    <>
      {groups
        .filter((g) => cit.group_ids.includes(g.id))
        .slice(0, 2)
        .map((g) => (
          <span key={g.id} className="pill" style={{ background: `${g.color}1A`, color: g.color, fontSize: 11, padding: '1px 8px' }}>
            {g.name}
          </span>
        ))}
    </>
  );
}
