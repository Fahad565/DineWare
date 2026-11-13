import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        safari: {
          50: '#faf8f5',
          100: '#f4ede2',
          200: '#e8dac3',
          300: '#d7c09e',
          400: '#c5a377',
          500: '#b48a58',
          600: '#9b7145',
          700: '#7d5838',
          800: '#664731',
          900: '#543b2a',
          950: '#2e1e14',
        },
        savanna: {
          50: '#fbfaf7',
          100: '#f5f3ec',
          200: '#e8e3d6',
          300: '#d6cdb7',
          400: '#beae92',
          500: '#a79273',
          600: '#8f795c',
          700: '#736049',
          800: '#5d4f3e',
          900: '#4e4235',
        },
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

export default config;
