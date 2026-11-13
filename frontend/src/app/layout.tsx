import type { Metadata } from 'next';
import './globals.css';
import Link from 'next/link';

export const metadata: Metadata = {
  title: 'DineWare Link | Safari Dietary & Itinerary Sync',
  description:
    'Offline-first dietary synchronization engine connecting safari tour operators with remote camp kitchens via Africa’s Talking SMS and USSD.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen flex flex-col font-sans antialiased text-slate-800 bg-stone-50">
        <header className="sticky top-0 z-40 bg-stone-900 text-stone-100 shadow-md border-b border-stone-800">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <Link href="/" className="flex items-center space-x-3 group">
                <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-amber-600 to-amber-700 flex items-center justify-center text-white shadow font-bold text-xl tracking-wider group-hover:scale-105 transition-transform">
                  DW
                </div>
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="font-bold text-lg text-white tracking-tight">DineWare Link</span>
                    <span className="text-xs bg-amber-500/20 text-amber-300 border border-amber-500/30 px-2 py-0.5 rounded-full font-medium">
                      Safari Operator
                    </span>
                  </div>
                  <p className="text-xs text-stone-400 hidden sm:block">
                    Zero-Data Kitchen Sync via Africa’s Talking GSM Gateway
                  </p>
                </div>
              </Link>
            </div>

            <nav className="flex items-center space-x-3 sm:space-x-4 text-sm font-medium">
              <Link
                href="/"
                className="px-3 py-1.5 rounded-md text-stone-300 hover:text-white hover:bg-stone-800 transition-colors"
              >
                Tour Operator Dashboard
              </Link>
              <Link
                href="/manifest/ITIN-2026-001"
                className="px-3 py-1.5 rounded-md text-amber-300 hover:text-amber-200 hover:bg-amber-950/50 border border-amber-500/30 transition-colors flex items-center space-x-1.5"
              >
                <span>Lodge Manifest</span>
                <span className="text-xs bg-amber-500/30 text-amber-200 px-1.5 py-0.2 rounded font-mono">
                  Lounge Wi-Fi
                </span>
              </Link>
            </nav>
          </div>
        </header>

        <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {children}
        </main>

        <footer className="bg-stone-900 border-t border-stone-800 text-stone-400 py-6 text-xs">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="flex items-center space-x-2">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              <span>Africa’s Talking Telephony Gateways: SMS Trigger &amp; USSD Shortcode <code className="bg-stone-800 px-1 py-0.5 rounded text-amber-400 font-mono">*384*44#</code></span>
            </div>
            <div className="text-stone-400">
              DineWare Link • Safari Tourism Hackathon • Powered by Go &amp; Next.js
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
