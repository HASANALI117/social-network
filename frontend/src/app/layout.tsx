import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { SidebarLayout } from "@/components/ui/sidebar-layout";
import { AppSidebar } from "@/components/layout/sidebar";
import { AppNavbar } from "@/components/layout/navbar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Social Network",
  description: "Stay connected with your friends and family.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="bg-white lg:bg-zinc-100 dark:bg-zinc-900 dark:lg:bg-zinc-950">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <SidebarLayout 
          navbar={<AppNavbar />}
          sidebar={<AppSidebar />}
        >
          {children}
        </SidebarLayout>
      </body>
    </html>
  );
}
