import { useState, useEffect } from 'react';
import { Outlet, NavLink } from 'react-router-dom';
import { useStats } from '../hooks/useStats';
import { useTheme } from '../hooks/useTheme';
import AskOverlay from './AskOverlay';

export default function Layout() {
  const { data: stats } = useStats();
  const { theme, toggleTheme } = useTheme();
  const [askOpen, setAskOpen] = useState(false);

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setAskOpen((v) => !v);
      }
    };
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      {/* Top bar */}
      <header
        style={{
          height: 48,
          borderBottom: '1px solid var(--color-border)',
          display: 'flex',
          alignItems: 'center',
          padding: '0 20px',
          gap: 16,
          flexShrink: 0,
        }}
      >
        <span
          style={{
            fontWeight: 590,
            fontSize: 14,
            color: 'var(--color-text-primary)',
          }}
        >
          laterbase
        </span>
        {stats !== undefined && (
          <span
            style={{
              fontSize: 12,
              color: 'var(--color-text-muted)',
              marginLeft: 4,
            }}
          >
            {stats.total_links} links
          </span>
        )}
        {stats !== undefined && stats.total_llm_cost > 0 && (
          <span
            style={{
              fontSize: 12,
              color: 'var(--color-text-muted)',
            }}
          >
            · ${stats.total_llm_cost.toFixed(4)} LLM
          </span>
        )}

        {/* Desktop nav */}
        <nav
          style={{
            marginLeft: 'auto',
            display: 'flex',
            gap: 4,
            alignItems: 'center',
          }}
        >
          <NavLink
            to="/"
            end
            style={({ isActive }) => navLinkStyle(isActive)}
          >
            Dashboard
          </NavLink>
          <NavLink
            to="/groups"
            style={({ isActive }) => navLinkStyle(isActive)}
          >
            Groups
          </NavLink>

          {/* Ask button */}
          <button
            onClick={() => setAskOpen(true)}
            title="Ask your archive (⌘K)"
            style={{
              marginLeft: 4,
              padding: '4px 12px',
              borderRadius: 'var(--radius-md)',
              border: '1px solid var(--color-border)',
              background: 'var(--color-surface)',
              color: 'var(--color-text-primary)',
              fontSize: 13,
              fontWeight: 510,
              cursor: 'pointer',
              lineHeight: 1.4,
              transition: 'background 0.15s ease, border-color 0.15s ease',
            }}
          >
            Ask <span style={{ color: 'var(--color-text-muted)', fontWeight: 400 }}>⌘K</span>
          </button>

          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}
            style={{
              marginLeft: 4,
              background: 'var(--color-surface)',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
              padding: '4px 8px',
              fontSize: 14,
              cursor: 'pointer',
              color: 'var(--color-text-secondary)',
              lineHeight: 1,
            }}
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
        </nav>
      </header>

      {/* Main content */}
      <main style={{ flex: 1, overflow: 'auto', padding: 20 }}>
        <Outlet />
      </main>

      {askOpen && <AskOverlay onClose={() => setAskOpen(false)} />}
    </div>
  );
}

function navLinkStyle(isActive: boolean): React.CSSProperties {
  return {
    padding: '4px 12px',
    borderRadius: 'var(--radius-md)',
    fontSize: 13,
    fontWeight: 510,
    textDecoration: 'none',
    color: isActive ? 'var(--color-text-primary)' : 'var(--color-text-secondary)',
    background: isActive ? 'var(--color-surface)' : 'transparent',
    transition: 'background 0.15s ease',
  };
}
