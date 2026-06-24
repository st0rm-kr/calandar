/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: '#F7F8FA',
        surface: '#FFFFFF',
        ink: '#14181F',
        muted: '#5B6472',
        hairline: '#E7E9EF',
        brand: {
          DEFAULT: '#2F6BFF',
          soft: '#EAF0FF',
          ink: '#1B45B8',
        },
        grass: {
          DEFAULT: '#13C37C',
          soft: '#E2F8EE',
        },
        tangerine: {
          DEFAULT: '#FF7A1A',
          soft: '#FFEEDF',
        },
        grape: {
          DEFAULT: '#8B5CF6',
          soft: '#F0EAFE',
        },
        rose: {
          DEFAULT: '#F43F5E',
          soft: '#FFE7EC',
        },
      },
      fontFamily: {
        display: ['Fredoka', 'system-ui', 'sans-serif'],
        sans: ['Nunito', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        soft: '0 8px 24px rgba(20, 24, 31, 0.08)',
        card: '0 6px 18px rgba(20, 24, 31, 0.06)',
        pop: '0 14px 30px rgba(47, 107, 255, 0.32)',
        tabbar: '0 10px 30px rgba(20, 24, 31, 0.12)',
      },
      borderRadius: {
        '4xl': '2rem',
      },
    },
  },
  plugins: [],
}
