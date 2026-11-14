import { defineTheme } from '@astryxdesign/core/theme';

type Token = string | [light: string, dark: string];

// Graphite & Lime (DESIGN.md section 2) mapped onto Astryx semantic token slots.
// Tuples are [light, dark]. Dark values are the proposed baseline (DES-04).
export const tokens: Record<string, Token> = {
  '--color-background-body': ['#F6F7F4', '#111710'],
  '--color-background-surface': ['#FFFFFF', '#1B241A'],
  '--color-background-popover': ['#FFFFFF', '#243021'],
  '--color-background-muted': ['#EDF2E7', '#2C3926'],
  '--color-text-primary': ['#1D2420', '#F4F7EF'],
  '--color-text-secondary': ['#576454', '#C1CCB9'],
  '--color-border': ['#DCE3D6', '#43503C'],
  '--color-accent': ['#4D721D', '#BDE874'],
  '--color-on-accent': ['#FFFFFF', '#16220E'],
  '--color-text-accent': ['#4D721D', '#BDE874'],
  '--color-icon-accent': ['#4D721D', '#BDE874'],
  '--focus-outline-color': ['#275DCA', '#9CC4FF'],
  '--color-text-success': ['#28683F', '#8CDEA0'],
  '--color-text-warning': ['#8A540D', '#F3C46F'],
  '--color-text-error': ['#B42318', '#FFB0A8'],
  '--font-family-body':
    'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  '--font-family-heading':
    'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  '--font-family-code': 'ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace',
};

export const theme226os = defineTheme({
  name: '226os',
  // DESIGN.md: 16px body text (Astryx core default is 14px). ratio 1.2 is the documented default.
  typography: { scale: { base: 16, ratio: 1.2 } },
  tokens,
  components: {
    // Astryx paints the bare track with the muted surface, so a bar at 0 % vanishes on a muted
    // card (the profile checklist). The border tone keeps the track visible on every surface.
    'progress-bar-track': { base: { backgroundColor: 'var(--color-border)' } },
  },
});
