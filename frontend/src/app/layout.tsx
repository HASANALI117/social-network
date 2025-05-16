'use client';

import type { Metadata } from 'next';
import { Geist, Geist_Mono } from 'next/font/google';
import '@/styles/globals.css';
import { SidebarLayout } from '@/components/ui/sidebar-layout';
import { AppSidebar } from '@/components/layout/sidebar';
import { AppNavbar } from '@/components/layout/navbar';
import { Toaster } from 'react-hot-toast';
import { useEffect } from 'react';
import { useUserStore } from '@/store/useUserStore';
import { GlobalWebSocketProvider } from '@/contexts/GlobalWebSocketContext';

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
});

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
});

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  // Handle store hydration at the root level
  useEffect(() => {
    useUserStore.persist.rehydrate();
  }, []);

  return (
    <html
      lang="en"
      className="bg-white lg:bg-zinc-100 dark:bg-zinc-900 dark:lg:bg-zinc-950"
    >
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <GlobalWebSocketProvider>
          <SidebarLayout navbar={<AppNavbar />} sidebar={<AppSidebar />}>
            {children}
          </SidebarLayout>
          <Toaster position="bottom-right" />
        </GlobalWebSocketProvider>
      </body>
    </html>
  );
}
