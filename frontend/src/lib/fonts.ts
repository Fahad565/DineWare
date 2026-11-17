import { Inter, Playfair_Display } from 'next/font/google';
import localFont from 'next/font/local';

export const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
});

export const playfairDisplay = Playfair_Display({
  subsets: ['latin'],
  weight: ['700'],
  display: 'swap',
  variable: '--font-playfair',
});

export const liberationMono = localFont({
  src: [
    {
      path: '../fonts/LiberationMono-Regular.ttf',
      weight: '400',
      style: 'normal',
    },
    {
      path: '../fonts/LiberationMono-Bold.ttf',
      weight: '700',
      style: 'normal',
    },
  ],
  display: 'swap',
  variable: '--font-liberation-mono',
});
