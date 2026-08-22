import { useState, useEffect, useCallback } from 'react';

type Theme = 'dark' | 'light';

function getStoredTheme(): Theme {
  try {
    const stored = localStorage.getItem('laterbase-theme');
    if (stored === 'dark' || stored === 'light') return stored;
  } catch {}
  return 'light';
}

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme);

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    try {
      localStorage.setItem('laterbase-theme', theme);
    } catch {}
  }, [theme]);

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => (prev === 'dark' ? 'light' : 'dark'));
  }, []);

  return { theme, toggleTheme };
}
