import type { Metadata } from "next";
import "./globals.css";
import Providers from "./providers";
import { Toaster } from "sonner";

export const metadata: Metadata = {
  title: "go-short | Production-grade URL Shortener",
  description: "High-performance URL shortener with rich analytics, QR codes, and custom slugs. Built with Go and Next.js.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased dark">
      <body className="min-h-full flex flex-col grain-overlay">
        <Providers>{children}</Providers>
        <Toaster position="top-right" theme="dark" toastOptions={{ className: "cyberpunk-toast" }} />
      </body>
    </html>
  );
}
