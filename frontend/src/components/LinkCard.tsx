import { useNavigate } from 'react-router-dom';
import type { Link, GroupWithCount } from '../types';
import StatusBadge from './StatusBadge';
import { useUpdateLink } from '../hooks/useLinks';

interface LinkCardProps {
  link: Link;
  groups: GroupWithCount[];
}

function timeAgo(dateStr: string): string {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diff = now - date;
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (seconds < 60) return 'just now';
  if (minutes < 60) return `${minutes}m`;
  if (hours < 24) return `${hours}h`;
  if (days === 1) return 'yesterday';
  if (days < 7) return `${days}d`;
  return new Date(dateStr).toLocaleDateString();
}

function getDomain(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return url;
  }
}

function getStatusSummary(link: Link): string | null {
  switch (link.status) {
    case 'ready':
      return link.summary;
    case 'fetching':
      return 'Fetching…';
    case 'processing':
      return 'Summarizing…';
    case 'inaccessible':
      return 'This site blocks automated access';
    case 'paywalled':
      return 'Behind a paywall — open in browser to read';
    case 'failed':
      return link.error_message
        ? link.error_message.slice(0, 120)
        : 'Failed to process';
    case 'new':
      return 'Queued…';
    default:
      return null;
  }
}

export default function LinkCard({ link, groups }: LinkCardProps) {
  const navigate = useNavigate();
  const updateLink = useUpdateLink();
  const interestingGroup = groups.find((g) => g.name.toLowerCase() === 'interesting');
  const isInteresting = interestingGroup ? link.group_ids.includes(interestingGroup.id) : false;
  const linkGroups = groups.filter((g) => link.group_ids.includes(g.id));
  const visibleGroups = linkGroups.slice(0, 3);
  const overflowCount = linkGroups.length - 3;

  const handleClick = () => navigate(`/links/${link.id}`);

  return (
    <div
      className="card-surface"
      style={{
        position: 'relative',
        cursor: 'pointer',
        overflow: 'hidden',
        display: 'flex',
        flexDirection: 'column',
      }}
      onClick={handleClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleClick();
        }
      }}
    >
      {/* Interesting toggle */}
      <button
        className="link-card-interesting-toggle"
        data-active={isInteresting}
        onClick={(e) => {
          e.stopPropagation();
          updateLink.mutate({ id: link.id, data: { is_interesting: !isInteresting } });
        }}
        disabled={updateLink.isPending}
        title={isInteresting ? 'Mark not interesting' : 'Mark interesting'}
        style={{
          position: 'absolute',
          top: 8,
          right: 8,
          zIndex: 2,
          width: 28,
          height: 28,
          borderRadius: 'var(--radius-md)',
          border: '1px solid var(--color-border)',
          background: 'var(--color-surface)',
          color: isInteresting ? 'var(--color-accent)' : 'var(--color-text-muted)',
          fontSize: 16,
          lineHeight: 1,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {isInteresting ? '★' : '☆'}
      </button>

      {/* OG Image */}
      {link.og_image_url && (
        <div
          style={{
            width: '100%',
            aspectRatio: '16 / 9',
            overflow: 'hidden',
            background: 'var(--color-surface)',
          }}
        >
          <img
            src={link.og_image_url}
            alt=""
            loading="lazy"
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
              display: 'block',
            }}
            onError={(e) => {
              (e.currentTarget as HTMLImageElement).style.display = 'none';
            }}
          />
        </div>
      )}

      {/* Content */}
      <div style={{ padding: 14, flex: 1, display: 'flex', flexDirection: 'column' }}>
        {/* Top row: favicon + domain + time */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            marginBottom: 6,
          }}
        >
          {link.favicon_url ? (
            <img
              src={link.favicon_url}
              alt=""
              style={{ width: 14, height: 14, borderRadius: 2 }}
            />
          ) : (
            <span style={{ fontSize: 13, color: 'var(--color-text-muted)' }}>◆</span>
          )}
          <span
            style={{
              fontSize: 11,
              fontWeight: 510,
              color: 'var(--color-text-secondary)',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              flex: 1,
            }}
          >
            {getDomain(link.url)}
          </span>
          <span
            style={{
              fontSize: 11,
              color: 'var(--color-text-muted)',
              whiteSpace: 'nowrap',
              flexShrink: 0,
            }}
          >
            {timeAgo(link.created_at)}
          </span>
        </div>

        {/* Title */}
        <div
          style={{
            fontSize: 15,
            fontWeight: 590,
            color: 'var(--color-text-primary)',
            lineHeight: 1.35,
            marginBottom: 4,
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            overflow: 'hidden',
          }}
        >
          {link.title || link.url}
        </div>

        {/* Summary / status line */}
        {getStatusSummary(link) && (
          <div
            style={{
              fontSize: 13,
              color: 'var(--color-text-secondary)',
              lineHeight: 1.4,
              marginBottom: 'auto',
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
            }}
          >
            {getStatusSummary(link)}
          </div>
        )}

        {/* Spacer to push badges to bottom */}
        {!getStatusSummary(link) && <div style={{ flex: 1 }} />}

        {/* Badges row */}
        <div
          style={{
            display: 'flex',
            gap: 4,
            alignItems: 'center',
            marginTop: 8,
            paddingTop: 8,
            borderTop: '1px solid var(--color-border)',
          }}
        >
          {visibleGroups.map((g) => (
            <span
              key={g.id}
              className="pill"
              style={{
                background: `${g.color}1A`,
                color: g.color,
              }}
            >
              {g.name}
            </span>
          ))}
          {overflowCount > 0 && (
            <span
              className="pill"
              style={{
                background: 'var(--color-surface)',
                color: 'var(--color-text-muted)',
              }}
            >
              +{overflowCount}
            </span>
          )}
          {linkGroups.length === 0 && link.source_type && (
            <span
              className="pill"
              style={{
                background: 'var(--color-surface)',
                color: 'var(--color-text-secondary)',
              }}
            >
              {link.source_type}
            </span>
          )}
          <div style={{ marginLeft: 'auto' }}>
            <StatusBadge status={link.status} />
          </div>
        </div>
      </div>
    </div>
  );
}
