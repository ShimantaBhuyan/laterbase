import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useLinks } from '../hooks/useLinks';
import { useGroups } from '../hooks/useGroups';
import { useStats } from '../hooks/useStats';
import LinkCard from '../components/LinkCard';
import GroupChips from '../components/GroupChips';
import AddLinkDialog from '../components/AddLinkDialog';
import EmptyState from '../components/EmptyState';

const PER_PAGE = 24;

const statusFilters = [
  { value: '', label: 'All' },
  { value: 'ready', label: 'Ready' },
  { value: 'fetching', label: 'Fetching' },
  { value: 'processing', label: 'Summarizing' },
  { value: 'inaccessible', label: 'Blocked' },
  { value: 'failed', label: 'Failed' },
];

function getPageNumbers(current: number, total: number): (number | 'ellipsis')[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }
  const pages: (number | 'ellipsis')[] = [];
  if (current <= 4) {
    for (let i = 1; i <= 5; i++) pages.push(i);
    pages.push('ellipsis');
    pages.push(total);
  } else if (current >= total - 3) {
    pages.push(1);
    pages.push('ellipsis');
    for (let i = total - 4; i <= total; i++) pages.push(i);
  } else {
    pages.push(1);
    pages.push('ellipsis');
    for (let i = current - 1; i <= current + 1; i++) pages.push(i);
    pages.push('ellipsis');
    pages.push(total);
  }
  return pages;
}

export default function Dashboard() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [showAddDialog, setShowAddDialog] = useState(false);

  const groupId = searchParams.get('group_id') || undefined;
  const status = searchParams.get('status') || undefined;
  const search = searchParams.get('search') || undefined;
  const page = parseInt(searchParams.get('page') || '1', 10);

  const [searchInput, setSearchInput] = useState(search || '');
  const [debouncedSearch, setDebouncedSearch] = useState(search || '');

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const offset = (page - 1) * PER_PAGE;

  const { data, isLoading } = useLinks({
    group_id: groupId,
    status,
    search: debouncedSearch || undefined,
    sort: 'created_at',
    order: 'desc',
    limit: PER_PAGE,
    offset,
  });

  const links = data?.links;
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PER_PAGE));

  const { data: groups } = useGroups();
  const { data: stats } = useStats();

  const updateParam = useCallback(
    (key: string, value: string | null) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (value) {
          next.set(key, value);
        } else {
          next.delete(key);
        }
        if (key !== 'page') {
          next.delete('page');
        }
        return next;
      });
    },
    [setSearchParams]
  );

  const setPage = useCallback(
    (p: number) => {
      if (p < 1 || p > totalPages) return;
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (p > 1) {
          next.set('page', String(p));
        } else {
          next.delete('page');
        }
        return next;
      });
    },
    [setSearchParams, totalPages]
  );

  const clearFilters = () => setSearchParams(new URLSearchParams());

  const hasFilters = !!(groupId || status || debouncedSearch);

  return (
    <div>
      {/* Heading row */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 16,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
          <h1
            style={{
              fontSize: 20,
              fontWeight: 590,
              color: 'var(--color-text-primary)',
              margin: 0,
            }}
          >
            Links
          </h1>
          {stats && (
            <span
              style={{
                fontSize: 13,
                color: 'var(--color-text-muted)',
              }}
            >
              {stats.total_links}
            </span>
          )}
        </div>
        <button
          onClick={() => setShowAddDialog(true)}
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
          Add Link
        </button>
      </div>

      {/* Search */}
      <input
        className="ghost"
        type="text"
        placeholder="Search links..."
        aria-label="Search links"
        value={searchInput}
        onChange={(e) => setSearchInput(e.target.value)}
        style={{
          width: '100%',
          padding: '10px 12px',
          fontSize: 14,
          marginBottom: 12,
          boxSizing: 'border-box',
        }}
      />

      {/* Group chips */}
      {groups && groups.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <GroupChips
            groups={groups}
            selected={groupId ?? null}
            onSelect={(id) => updateParam('group_id', id)}
          />
        </div>
      )}

      {/* Status filter pills */}
      <div
        style={{
          display: 'flex',
          gap: 6,
          marginBottom: 16,
          flexWrap: 'wrap',
        }}
      >
        {statusFilters.map((sf) => {
          const isActive = (sf.value === '' && !status) || sf.value === status;
          return (
            <button
              key={sf.value}
              onClick={() => {
                updateParam('status', sf.value || null);
              }}
              style={{
                padding: '4px 14px',
                borderRadius: 'var(--radius-pill)',
                border: '1px solid',
                borderColor: isActive
                  ? 'var(--color-accent)'
                  : 'var(--color-border)',
                background: isActive ? 'var(--color-accent)' : 'transparent',
                color: isActive ? '#fff' : 'var(--color-text-secondary)',
                fontSize: 12,
                fontWeight: 510,
                cursor: 'pointer',
              }}
            >
              {sf.label}
            </button>
          );
        })}
      </div>

      {/* Loading skeletons */}
      {isLoading ? (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
            gap: 12,
          }}
        >
          {Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="card-surface"
              style={{
                padding: 0,
                overflow: 'hidden',
                opacity: 0.3,
                animation: 'pulse 1.5s infinite',
              }}
            >
              <div
                style={{
                  aspectRatio: '16 / 9',
                  background: 'var(--color-border)',
                }}
              />
              <div style={{ padding: 14 }}>
                <div
                  style={{
                    height: 10,
                    width: '50%',
                    background: 'var(--color-border)',
                    borderRadius: 4,
                    marginBottom: 8,
                  }}
                />
                <div
                  style={{
                    height: 16,
                    width: '85%',
                    background: 'var(--color-border)',
                    borderRadius: 4,
                    marginBottom: 6,
                  }}
                />
                <div
                  style={{
                    height: 12,
                    width: '60%',
                    background: 'var(--color-border)',
                    borderRadius: 4,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      ) : links && links.length > 0 ? (
        <>
          {/* Grid */}
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))',
              gap: 12,
            }}
          >
            {links.map((link) => (
              <LinkCard key={link.id} link={link} groups={groups ?? []} />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 4,
                marginTop: 24,
                paddingTop: 16,
                borderTop: '1px solid var(--color-border)',
              }}
            >
              <button
                onClick={() => setPage(page - 1)}
                disabled={page <= 1}
                style={paginationBtnStyle(page <= 1)}
                aria-label="Previous page"
              >
                ←
              </button>
              {getPageNumbers(page, totalPages).map((p, i) =>
                p === 'ellipsis' ? (
                  <span
                    key={`e-${i}`}
                    style={{
                      fontSize: 12,
                      color: 'var(--color-text-muted)',
                      padding: '0 4px',
                    }}
                  >
                    …
                  </span>
                ) : (
                  <button
                    key={p}
                    onClick={() => setPage(p)}
                    style={{
                      ...paginationBtnStyle(p === page),
                      fontWeight: p === page ? 590 : 400,
                      background:
                        p === page
                          ? 'var(--color-surface-elevated)'
                          : 'transparent',
                      color:
                        p === page
                          ? 'var(--color-text-primary)'
                          : 'var(--color-text-secondary)',
                    }}
                  >
                    {p}
                  </button>
                )
              )}
              <button
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages}
                style={paginationBtnStyle(page >= totalPages)}
                aria-label="Next page"
              >
                →
              </button>
            </div>
          )}

          {/* Results info */}
          <div
            style={{
              textAlign: 'center',
              fontSize: 11,
              color: 'var(--color-text-muted)',
              marginTop: 8,
            }}
          >
            {offset + 1}–{Math.min(offset + PER_PAGE, total)} of {total}
          </div>
        </>
      ) : (
        <EmptyState hasFilters={hasFilters} onClearFilters={clearFilters} />
      )}

      {showAddDialog && <AddLinkDialog onClose={() => setShowAddDialog(false)} />}
    </div>
  );
}

function paginationBtnStyle(
  disabled: boolean
): React.CSSProperties {
  return {
    padding: '4px 10px',
    borderRadius: 'var(--radius-md)',
    border: 'none',
    fontSize: 13,
    cursor: disabled ? 'default' : 'pointer',
    opacity: disabled ? 0.3 : 1,
    lineHeight: 1.4,
    minWidth: 32,
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: 'background 0.15s ease, color 0.15s ease',
  };
}
