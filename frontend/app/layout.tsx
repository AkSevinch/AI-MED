import type { Metadata } from "next";
import "./globals.css";
import { SidebarProvider } from "@/components/ui/sidebar";
import AppShell from "@/components/custom/AppShell";
import AuthGate from "@/components/custom/AuthGate";

export const metadata: Metadata = {
  title: "AI MED — единая сводка здоровья",
  description:
    "Собирайте анализы, измерения и историю здоровья в одну понятную сводку.",
  icons: {
    icon: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='white'%3E%3Cg transform='rotate(-30 12 12)'%3E%3Ccircle cx='7.3' cy='3.2' r='1.45'/%3E%3Crect x='5.5' y='4.7' width='3.6' height='14.6' rx='1.8'/%3E%3Crect x='14.9' y='4.7' width='3.6' height='14.6' rx='1.8'/%3E%3Ccircle cx='16.7' cy='20.8' r='1.45'/%3E%3C/g%3E%3C/svg%3E",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru" suppressHydrationWarning>
      <body>
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){try{var t=localStorage.getItem('aimed-theme');var d=t?t==='dark':window.matchMedia('(prefers-color-scheme: dark)').matches;if(d)document.documentElement.classList.add('dark');}catch(e){}})();`,
          }}
        />
        <SidebarProvider>
          <AuthGate>
            <AppShell>{children}</AppShell>
          </AuthGate>
        </SidebarProvider>
      </body>
    </html>
  );
}
