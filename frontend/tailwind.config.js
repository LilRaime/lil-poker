/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontSize: {
        'xxs': '0.625rem',
        '3xs': '0.5rem',
        'xxxxs': '0.375rem',
      },
      spacing: {
        '18': '4.5rem',
      }
    },
  },
  plugins: [],
}