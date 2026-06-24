/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        canvas: '#050507',
        surface: '#101219',
        ink: '#F8FAFC',
        muted: '#A3AAB8',
        hairline: '#252A36',
        brand: {
          DEFAULT: '#00F2EA',
          soft: '#073236',
          ink: '#7FFCF7',
        },
        grass: {
          DEFAULT: '#3DFF88',
          soft: '#0A2E1B',
        },
        tangerine: {
          DEFAULT: '#FF7A1A',
          soft: '#3A2108',
        },
        grape: {
          DEFAULT: '#9B5CFF',
          soft: '#23153D',
        },
        rose: {
          DEFAULT: '#FF0050',
          soft: '#3A0719',
        },
      },
      fontFamily: {
        display: ['Fredoka', 'system-ui', 'sans-serif'],
        sans: ['Nunito', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        soft: '0 18px 60px rgba(0, 0, 0, 0.32)',
        card: '0 14px 44px rgba(0, 0, 0, 0.28)',
        pop: '0 0 28px rgba(0, 242, 234, 0.35), 0 0 42px rgba(255, 0, 80, 0.25)',
        tabbar: '0 20px 70px rgba(0, 0, 0, 0.45)',
      },
      borderRadius: {
        '4xl': '2rem',
      },
    },
  },
  plugins: [],
}
