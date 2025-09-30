/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./web/templates/**/*.html",
    "./web/static/js/**/*.js"
  ],
  theme: {
    extend: {
      colors: {
        // Eye in the Sky Dark Theme
        background: '#1B1F28',
        surface: '#242B38',

        // Text colors
        'text-primary': '#E5E9F0',
        'text-secondary': '#A0A9B8',

        // Status colors
        status: {
          active: '#3DDC97',
          idle: '#E0B95B',
          failed: '#F96363',
          completed: '#9CA3AF',
        },

        // Accent color
        accent: '#83AAFF',

        // Primary palette (for buttons/highlights)
        primary: {
          50: '#F0F5FF',
          100: '#E0EBFF',
          200: '#C7D9FF',
          300: '#A3C1FF',
          400: '#83AAFF',
          500: '#6B94FF',
          600: '#5578E8',
          700: '#4056C9',
          800: '#2D3FA3',
          900: '#1E2B75'
        }
      },
      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '112': '28rem',
        '128': '32rem'
      },
      fontFamily: {
        sans: ['Inter', 'Roboto', 'system-ui', '-apple-system', 'sans-serif']
      },
      letterSpacing: {
        'wider': '0.05em',
        'widest': '0.1em'
      },
      animation: {
        'pulse-slow': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite'
      },
      boxShadow: {
        'glow-sm': '0 0 10px rgba(131, 170, 255, 0.3)',
        'glow': '0 0 20px rgba(131, 170, 255, 0.4)',
        'glow-lg': '0 0 30px rgba(131, 170, 255, 0.5)',
      }
    }
  },
  plugins: []
}