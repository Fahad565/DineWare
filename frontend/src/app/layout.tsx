import type { Metadata } from 'next';
import './globals.css';
import Link from 'next/link';
import { inter, liberationMono, playfairDisplay } from '../lib/fonts';

export const metadata: Metadata = {
  title: 'DineWare | Safari operations',
  description:
    'Keep safari itineraries and lodge kitchen requirements in sync.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${playfairDisplay.variable} ${liberationMono.variable}`}
    >
      <body className="min-h-screen flex flex-col font-sans antialiased text-stone-800 bg-stone-50">
        <header className="sticky top-0 z-40 border-b border-stone-200 bg-stone-50/95 text-stone-900 backdrop-blur">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <Link href="/" className="flex items-center space-x-3 group">
                <div className="w-9 h-9 rounded-xl bg-amber-700 flex items-center justify-center text-white font-semibold text-sm tracking-wide font-mono">
                  DW
                </div>
                <div>
                  <span className="font-display text-base font-bold tracking-tight text-stone-900">DineWare</span>
                  <p className="text-xs text-stone-500 hidden sm:block">Safari operations</p>
                </div>
              </Link>
            </div>

            <nav className="flex items-center gap-1 text-sm font-medium">
              <Link
                href="/"
                className="px-3 py-2 rounded-lg text-stone-700 hover:text-stone-950 hover:bg-stone-100 transition-colors"
              >
                Dashboard
              </Link>
              <Link
                href="/manifest/ITIN-2026-001"
                className="px-3 py-2 rounded-lg text-stone-700 hover:text-stone-950 hover:bg-stone-100 transition-colors"
              >
                Lodge manifest
              </Link>
            </nav>
          </div>
        </header>

        <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {children}
        </main>

        <footer className="border-t border-stone-200 bg-white text-stone-500 py-5 text-xs">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="flex items-center space-x-2">
              <span className="w-2 h-2 rounded-full bg-emerald-600"></span>
              <span>Kitchen updates work over SMS and USSD <code className="bg-stone-100 px-1.5 py-0.5 rounded text-stone-700 font-mono">*384*44#</code></span>
            </div>
            <div>
              DineWare · Safari Tourism Hackathon
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
