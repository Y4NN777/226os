import { Theme } from '@astryxdesign/core/theme';
import {
  createContext,
  useContext,
  useMemo,
  useState,
  useSyncExternalStore,
  type ReactNode,
} from 'react';
import { readStorage, writeStorage } from '../app/storage';
import { theme226os } from './theme226os';

export type ThemeMode = 'light' | 'dark' | 'system';

const STORAGE_KEY = '226os.theme';
const modes: readonly ThemeMode[] = ['light', 'dark', 'system'];

interface ThemePreference {
  mode: ThemeMode;
  setMode: (mode: ThemeMode) => void;
}

const ThemePreferenceContext = createContext<ThemePreference | null>(null);

function initialMode(): ThemeMode {
  const stored = readStorage(STORAGE_KEY);
  return modes.find((mode) => mode === stored) ?? 'system';
}

/** One Astryx Theme for the whole app, with an explicit light/dark/system preference. */
export function ThemePreferenceProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(initialMode);
  const value = useMemo<ThemePreference>(
    () => ({
      mode,
      setMode: (next) => {
        writeStorage(STORAGE_KEY, next);
        setModeState(next);
      },
    }),
    [mode],
  );
  return (
    <ThemePreferenceContext value={value}>
      <Theme theme={theme226os} mode={mode}>
        {children}
      </Theme>
    </ThemePreferenceContext>
  );
}

export function useThemePreference(): ThemePreference {
  const value = useContext(ThemePreferenceContext);
  if (!value) throw new Error('useThemePreference must be used inside <ThemePreferenceProvider>');
  return value;
}

const DARK_QUERY = '(prefers-color-scheme: dark)';

function darkQuery(): MediaQueryList | null {
  return typeof window.matchMedia === 'function' ? window.matchMedia(DARK_QUERY) : null;
}

function subscribeToSystemTheme(onChange: () => void) {
  const query = darkQuery();
  query?.addEventListener('change', onChange);
  return () => query?.removeEventListener('change', onChange);
}

function systemPrefersDark() {
  return darkQuery()?.matches ?? false;
}

/** The theme actually shown: the stored choice, or the system one when the choice is "system". */
export function useResolvedTheme(): 'light' | 'dark' {
  const { mode } = useThemePreference();
  const systemDark = useSyncExternalStore(subscribeToSystemTheme, systemPrefersDark, () => false);
  if (mode !== 'system') return mode;
  return systemDark ? 'dark' : 'light';
}
