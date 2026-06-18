/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'Segoe UI', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Cascadia Mono', 'Consolas', 'monospace'],
      },
      colors: {
        ink: {
          950: '#09090b',
          900: '#0f1115',
          850: '#151821',
          800: '#1b1f2a',
          750: '#222735',
          700: '#2b3140',
        },
      },
      boxShadow: {
        panel: '0 18px 60px rgba(0,0,0,.28)',
      },
    },
  },
  plugins: [
    function ({ addVariant }) {
      addVariant('light', '.light &');
    },
  ],
};
