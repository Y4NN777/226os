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
  // Input, select and textarea outlines: the strong border (3:1 or more on every surface).
  '--color-border-emphasized': ['#6F7C6B', '#7E8E75'],
  '--color-accent': ['#4D721D', '#BDE874'],
  '--color-on-accent': ['#FFFFFF', '#16220E'],
  '--color-text-accent': ['#4D721D', '#BDE874'],
  '--color-icon-accent': ['#4D721D', '#BDE874'],
  '--focus-outline-color': ['#275DCA', '#9CC4FF'],
  '--color-text-success': ['#28683F', '#8CDEA0'],
  '--color-text-warning': ['#8A540D', '#F3C46F'],
  '--color-text-error': ['#B42318', '#FFB0A8'],
  '--color-text-info': ['#1F4FA8', '#9CC4FF'],
  // Status fills and their text (DESIGN.md 2.1; pairs audited in prototypes/html tokens.css).
  '--color-success-muted': ['#E3F1E7', '#1B3324'],
  '--color-on-success-muted': ['#28683F', '#8CDEA0'],
  '--color-info-muted': ['#E4ECFA', '#1D2F4D'],
  '--color-on-info-muted': ['#1F4FA8', '#9CC4FF'],
  '--color-warning-muted': ['#FBEFD9', '#3A2D12'],
  '--color-on-warning-muted': ['#8A540D', '#F3C46F'],
  '--color-error-muted': ['#FCE8E5', '#3E1F1C'],
  '--color-on-error-muted': ['#B42318', '#FFB0A8'],
  '--color-icon-secondary': ['#576454', '#C1CCB9'],
  '--color-icon-success': ['#28683F', '#8CDEA0'],
  '--color-icon-info': ['#1F4FA8', '#9CC4FF'],
  '--color-icon-warning': ['#8A540D', '#F3C46F'],
  '--color-icon-error': ['#B42318', '#FFB0A8'],
  '--color-skeleton': ['#E3EADC', '#2C3926'],
  // Control heights (prototype tokens.css: --target 44, .btn-sm 36, .search-field-lg 52). Astryx
  // reads these for buttons, inputs, selectors and tabs, so one change sets every page.
  '--size-element-sm': '36px',
  '--size-element-md': '44px',
  '--size-element-lg': '52px',
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
