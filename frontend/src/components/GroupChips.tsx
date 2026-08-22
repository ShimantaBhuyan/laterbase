import type { GroupWithCount } from '../types';

interface GroupChipsProps {
  groups: GroupWithCount[];
  selected: string | null;
  onSelect: (id: string | null) => void;
}

export default function GroupChips({
  groups,
  selected,
  onSelect,
}: GroupChipsProps) {
  return (
    <div
      style={{
        display: 'flex',
        gap: 6,
        overflowX: 'auto',
        paddingBottom: 4,
        WebkitOverflowScrolling: 'touch',
      }}
      className="hide-scrollbar"
      role="tablist"
      aria-label="Filter by group"
    >
      {/* All chip */}
      <button
        onClick={() => onSelect(null)}
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 6,
          padding: '4px 14px',
          borderRadius: 'var(--radius-pill)',
          border: '1px solid',
          borderColor:
            selected === null
              ? 'var(--color-accent)'
              : 'var(--color-border)',
          background:
            selected === null
              ? 'var(--color-accent)'
              : 'transparent',
          color:
            selected === null
              ? '#fff'
              : 'var(--color-text-secondary)',
          fontSize: 12,
          fontWeight: 510,
          cursor: 'pointer',
          whiteSpace: 'nowrap',
          flexShrink: 0,
        }}
      >
        All
      </button>

      {groups.map((group) => {
        const isSelected = selected === group.id;
        return (
          <button
            key={group.id}
            onClick={() => onSelect(group.id)}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '4px 14px',
              borderRadius: 'var(--radius-pill)',
              border: '1px solid',
              borderColor: isSelected ? group.color : 'var(--color-border)',
              background: isSelected ? `${group.color}26` : 'transparent',
              color: isSelected ? group.color : 'var(--color-text-secondary)',
              fontSize: 12,
              fontWeight: 510,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              flexShrink: 0,
            }}
          >
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                background: group.color,
                display: 'inline-block',
                flexShrink: 0,
              }}
            />
            {group.name}
            <span
              style={{
                fontSize: 11,
                opacity: 0.7,
              }}
            >
              {group.link_count}
            </span>
          </button>
        );
      })}
    </div>
  );
}
