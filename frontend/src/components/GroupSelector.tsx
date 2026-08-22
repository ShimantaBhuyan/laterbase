import { useState } from 'react';
import { useUpdateLink } from '../hooks/useLinks';
import type { GroupWithCount } from '../types';

interface GroupSelectorProps {
  linkId: string;
  currentGroupIds: string[];
  groups: GroupWithCount[];
}

const MAX_VISIBLE = 4;

export default function GroupSelector({
  linkId,
  currentGroupIds,
  groups,
}: GroupSelectorProps) {
  const [editing, setEditing] = useState(false);
  const updateLink = useUpdateLink();

  const selected = new Set(currentGroupIds);

  const toggle = (gid: string) => {
    const next = selected.has(gid)
      ? currentGroupIds.filter((id) => id !== gid)
      : [...currentGroupIds, gid];
    updateLink.mutate(
      { id: linkId, data: { group_ids: next } },
      { onSuccess: () => setEditing(false) }
    );
  };

  const selectedGroups = groups.filter((g) => selected.has(g.id));
  const visibleDisplay = selectedGroups.slice(0, MAX_VISIBLE);
  const overflowDisplay = selectedGroups.length - MAX_VISIBLE;

  if (editing) {
    return (
      <div
        style={{
          padding: '6px 10px',
          borderRadius: 'var(--radius-md)',
          border: '1px solid var(--color-border)',
          background: 'var(--color-surface)',
        }}
      >
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 4,
          }}
        >
          {groups.map((g) => {
            const isOn = selected.has(g.id);
            return (
              <button
                key={g.id}
                onClick={() => toggle(g.id)}
                disabled={updateLink.isPending}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 4,
                  padding: '3px 10px',
                  borderRadius: 'var(--radius-pill)',
                  border: '1px solid',
                  borderColor: isOn ? g.color : 'var(--color-border)',
                  background: isOn ? `${g.color}1A` : 'transparent',
                  color: isOn ? g.color : 'var(--color-text-secondary)',
                  fontSize: 12,
                  fontWeight: 510,
                  cursor: updateLink.isPending ? 'not-allowed' : 'pointer',
                  opacity: updateLink.isPending ? 0.6 : 1,
                  fontFamily: 'inherit',
                }}
              >
                <span
                  style={{
                    width: 7,
                    height: 7,
                    borderRadius: '50%',
                    background: isOn ? g.color : 'var(--color-border)',
                    display: 'inline-block',
                    flexShrink: 0,
                  }}
                />
                {g.name}
              </button>
            );
          })}
        </div>
        <div style={{ marginTop: 6, display: 'flex', gap: 6 }}>
          <button
            onClick={() => setEditing(false)}
            style={{
              padding: '2px 8px',
              borderRadius: 'var(--radius-sm)',
              border: 'none',
              background: 'transparent',
              color: 'var(--color-text-muted)',
              fontSize: 11,
              cursor: 'pointer',
              fontFamily: 'inherit',
            }}
          >
            Done
          </button>
        </div>
      </div>
    );
  }

  return (
    <div
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        flexWrap: 'wrap',
      }}
    >
      {selectedGroups.length === 0 ? (
        <span
          style={{
            fontSize: 12,
            color: 'var(--color-text-muted)',
            fontWeight: 510,
          }}
        >
          Uncategorized
        </span>
      ) : (
        visibleDisplay.map((g) => (
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
        ))
      )}
      {overflowDisplay > 0 && (
        <span
          style={{
            fontSize: 11,
            color: 'var(--color-text-muted)',
            fontWeight: 510,
          }}
        >
          +{overflowDisplay}
        </span>
      )}
      <button
        onClick={() => setEditing(true)}
        style={{
          background: 'none',
          border: 'none',
          color: 'var(--color-accent)',
          fontSize: 12,
          cursor: 'pointer',
          padding: 0,
          fontFamily: 'inherit',
        }}
      >
        Edit
      </button>
    </div>
  );
}
