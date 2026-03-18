/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx,js,jsx}'],
  theme: {
    extend: {
      colors: {
        app: 'var(--bg-app)',
        sidebar: 'var(--bg-sidebar)',
        surface: {
          1: 'var(--bg-surface-1)',
          2: 'var(--bg-surface-2)',
          3: 'var(--bg-surface-3)',
        },
        border: {
          DEFAULT: 'var(--border-default)',
          muted: 'var(--border-muted)',
          card: 'var(--border-card)',
        },
        text: {
          primary: 'var(--text-primary)',
          secondary: 'var(--text-secondary)',
          tertiary: 'var(--text-tertiary)',
          inverse: 'var(--text-inverse)',
        },
        state: {
          success: 'var(--state-success)',
          warning: 'var(--state-warning)',
          danger: 'var(--state-danger)',
          info: 'var(--state-info)',
          neutral: 'var(--state-neutral)',
        },
        focus: {
          ring: 'var(--focus-ring)',
        },
      },
      fontFamily: {
        sans: ['var(--font-ui)'],
        heading: ['var(--font-head)'],
        mono: ['var(--font-mono)'],
      },
      borderRadius: {
        card: 'var(--radius-card)',
        panel: 'var(--radius-panel)',
        pill: 'var(--radius-pill)',
      },
      maxWidth: {
        shell: '1440px',
      },
      keyframes: {
        shimmer: {
          '0%': { backgroundPosition: '200% 0' },
          '100%': { backgroundPosition: '-200% 0' },
        },
      },
      animation: {
        shimmer: 'shimmer 1.8s ease-in-out infinite',
      },
    },
  },
  plugins: [],
}
