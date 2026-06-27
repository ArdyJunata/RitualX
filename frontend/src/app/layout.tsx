import type { Metadata } from "next";
import { Inter, Outfit } from "next/font/google";
import "./globals.css";
import QueryProvider from "@/shared/components/providers/QueryProvider";
import { AnalyticsProvider } from "@/shared/components/providers/AnalyticsProvider";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const outfit = Outfit({ subsets: ["latin"], variable: "--font-outfit" });

export const metadata: Metadata = {
  title: "RitualX",
  description: "Gamified habit tracker",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${inter.variable} ${outfit.variable} font-sans antialiased`}>
        <QueryProvider>
          <AnalyticsProvider>
            {children}
          </AnalyticsProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
