/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        kumi: {
          purple: '#8F00FF',
          'purple-dark': '#280047',
          'purple-bg': '#3B0963',
          dark: '#111',
          'dark-secondary': '#18141B',
          'dark-tertiary': '#212121',
          gray: '#777',
          'gray-light': '#959595',
        },
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'monospace'],
        rounded: ['M PLUS Rounded 1c', 'sans-serif'],
      },
      backgroundImage: {
        topo: "url('https://api.builder.io/api/v1/image/assets/TEMP/2586000e7d564a5ef576041b9cd5008126c2ff9b?width=1500')",
      },
    },
  },
  plugins: [],
}
