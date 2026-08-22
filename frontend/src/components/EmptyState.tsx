interface EmptyStateProps {
  hasFilters: boolean;
  onClearFilters?: () => void;
}

export default function EmptyState({ hasFilters, onClearFilters }: EmptyStateProps) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '60px 20px',
        textAlign: 'center',
      }}
    >
      <div
        aria-hidden="true"
        style={{
          fontSize: 48,
          marginBottom: 16,
          opacity: 0.3,
        }}
      >
        {hasFilters ? '🔍' : '🔗'}
      </div>
      <p
        style={{
          fontSize: 14,
          fontWeight: 400,
          color: 'var(--color-text-secondary)',
          margin: 0,
          lineHeight: 1.5,
        }}
      >
        {hasFilters
          ? 'No links match your filters'
          : 'Share a link via WhatsApp — send link <url>'}
      </p>
      {hasFilters && onClearFilters && (
        <button
          onClick={onClearFilters}
          style={{
            marginTop: 12,
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
          Clear filters
        </button>
      )}
    </div>
  );
}
