interface StatusBadgeProps {
  status: string;
}

const statusConfig: Record<
  string,
  { label: string; color: string }
> = {
  ready: { label: 'Ready', color: 'var(--color-green)' },
  fetching: { label: 'Fetching', color: 'var(--color-amber)' },
  processing: { label: 'Summarizing', color: 'var(--color-amber)' },
  inaccessible: { label: 'Blocked', color: 'var(--color-amber)' },
  paywalled: { label: 'Paywalled', color: 'var(--color-red)' },
  failed: { label: 'Failed', color: 'var(--color-red)' },
  new: { label: 'New', color: 'var(--color-text-muted)' },
};

export default function StatusBadge({ status }: StatusBadgeProps) {
  const cfg = statusConfig[status] ?? {
    label: status,
    color: 'var(--color-text-muted)',
  };

  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
      <span
        className="status-dot"
        style={{
          background: cfg.color,
          width: 8,
          height: 8,
          borderRadius: '50%',
          display: 'inline-block',
          flexShrink: 0,
        }}
      />
      <span
        style={{
          fontSize: 12,
          fontWeight: 510,
          color: 'var(--color-text-secondary)',
        }}
      >
        {cfg.label}
      </span>
    </span>
  );
}
