import { useState } from 'react';
import {
  useGroups,
  useCreateGroup,
  useUpdateGroup,
  useDeleteGroup,
} from '../hooks/useGroups';
import ColorPicker from '../components/ColorPicker';

export default function Groups() {
  const { data: groups, isLoading } = useGroups();
  const createGroup = useCreateGroup();
  const updateGroup = useUpdateGroup();
  const deleteGroup = useDeleteGroup();

  const [showNewForm, setShowNewForm] = useState(false);
  const [newName, setNewName] = useState('');
  const [newColor, setNewColor] = useState('#5e6ad2');

  // Inline editing state
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingName, setEditingName] = useState('');

  const handleCreate = async () => {
    if (!newName.trim()) return;
    try {
      await createGroup.mutateAsync({
        name: newName.trim(),
        color: newColor,
      });
      setNewName('');
      setNewColor('#5e6ad2');
      setShowNewForm(false);
    } catch {
      // error handled by react-query
    }
  };

  const handleDelete = (id: string, name: string) => {
    if (window.confirm(`Delete group "${name}"?`)) {
      deleteGroup.mutate(id);
    }
  };

  const handleEditName = async (id: string) => {
    if (!editingName.trim()) return;
    await updateGroup.mutateAsync({ id, data: { name: editingName.trim() } });
    setEditingId(null);
    setEditingName('');
  };

  return (
    <div style={{ maxWidth: 600, margin: '0 auto' }}>
      {/* Heading */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 16,
        }}
      >
        <h1
          style={{
            fontSize: 20,
            fontWeight: 590,
            color: 'var(--color-text-primary)',
            margin: 0,
          }}
        >
          Groups
        </h1>
        <button
          onClick={() => setShowNewForm(!showNewForm)}
          style={{
            padding: '8px 16px',
            borderRadius: 'var(--radius-md)',
            border: 'none',
            background: 'var(--color-accent)',
            color: '#fff',
            fontSize: 13,
            fontWeight: 510,
            cursor: 'pointer',
          }}
        >
          New Group
        </button>
      </div>

      {/* New group form */}
      {showNewForm && (
        <div
          className="card-surface"
          style={{ padding: 16, marginBottom: 12 }}
        >
          <div style={{ marginBottom: 12 }}>
            <label
              style={{
                fontSize: 12,
                fontWeight: 510,
                color: 'var(--color-text-secondary)',
                display: 'block',
                marginBottom: 4,
              }}
            >
              Name
            </label>
            <input
              className="ghost"
              type="text"
              placeholder="Group name"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreate();
              }}
              style={{
                width: '100%',
                padding: '8px 10px',
                fontSize: 13,
                boxSizing: 'border-box',
              }}
              autoFocus
            />
          </div>
          <div style={{ marginBottom: 12 }}>
            <label
              style={{
                fontSize: 12,
                fontWeight: 510,
                color: 'var(--color-text-secondary)',
                display: 'block',
                marginBottom: 4,
              }}
            >
              Color
            </label>
            <ColorPicker value={newColor} onChange={setNewColor} />
          </div>
          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <button
              onClick={() => setShowNewForm(false)}
              style={{
                padding: '6px 14px',
                borderRadius: 'var(--radius-md)',
                border: 'none',
                background: 'var(--color-surface)',
                color: 'var(--color-text-secondary)',
                fontSize: 12,
                fontWeight: 510,
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
            <button
              onClick={handleCreate}
              disabled={createGroup.isPending || !newName.trim()}
              style={{
                padding: '6px 14px',
                borderRadius: 'var(--radius-md)',
                border: 'none',
                background: 'var(--color-accent)',
                color: '#fff',
                fontSize: 12,
                fontWeight: 510,
                cursor: createGroup.isPending ? 'not-allowed' : 'pointer',
                opacity: createGroup.isPending || !newName.trim() ? 0.5 : 1,
              }}
            >
              {createGroup.isPending ? 'Creating…' : 'Create'}
            </button>
          </div>
        </div>
      )}

      {/* Group list */}
      {isLoading ? (
        <div style={{ color: 'var(--color-text-muted)', fontSize: 14 }}>
          Loading…
        </div>
      ) : groups && groups.length > 0 ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {groups.map((group) => (
            <div
              key={group.id}
              className="card-surface"
              style={{
                padding: '12px 16px',
                display: 'flex',
                alignItems: 'center',
                gap: 12,
              }}
            >
              {/* Color picker */}
              <ColorPicker
                value={group.color}
                onChange={(color) =>
                  updateGroup.mutate({ id: group.id, data: { color } })
                }
              />

              {/* Editable name */}
              <div style={{ flex: 1, minWidth: 0 }}>
                {editingId === group.id ? (
                  <input
                    className="ghost"
                    type="text"
                    value={editingName}
                    onChange={(e) => setEditingName(e.target.value)}
                    onBlur={() => handleEditName(group.id)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') handleEditName(group.id);
                      if (e.key === 'Escape') setEditingId(null);
                    }}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      fontSize: 13,
                      fontWeight: 510,
                      boxSizing: 'border-box',
                    }}
                    autoFocus
                  />
                ) : (
                  <span
                    onClick={() => {
                      setEditingId(group.id);
                      setEditingName(group.name);
                    }}
                    style={{
                      fontSize: 13,
                      fontWeight: 510,
                      color: 'var(--color-text-primary)',
                      cursor: 'pointer',
                    }}
                  >
                    {group.name}
                  </span>
                )}
              </div>

              {/* Link count */}
              <span
                style={{
                  fontSize: 12,
                  fontWeight: 510,
                  color: 'var(--color-text-muted)',
                  background: 'var(--color-surface)',
                  padding: '2px 8px',
                  borderRadius: 'var(--radius-pill)',
                }}
              >
                {group.link_count}
              </span>

              {/* Delete */}
              <button
                onClick={() => handleDelete(group.id, group.name)}
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'var(--color-text-muted)',
                  fontSize: 14,
                  cursor: 'pointer',
                  padding: 8,
                  opacity: 0.5,
                  transition: 'opacity 0.15s',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.opacity = '1')}
                onMouseLeave={(e) => (e.currentTarget.style.opacity = '0.5')}
                title="Delete group"
                aria-label={`Delete group ${group.name}`}
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      ) : (
        <div
          style={{
            color: 'var(--color-text-muted)',
            fontSize: 14,
            textAlign: 'center',
            padding: 40,
          }}
        >
          No groups yet. Create one to organize your links.
        </div>
      )}
    </div>
  );
}
