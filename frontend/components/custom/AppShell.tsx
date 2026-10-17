"use client";

import { usePathname } from "next/navigation";
import { AppSidebar } from "@/components/custom/app-sidebar";
import Header from "@/components/custom/Header";
import Footer from "@/components/custom/Footer";
import { SidebarInset } from "@/components/ui/sidebar";
import { useAuth } from "@/components/custom/AuthGate";

const BARE_PATHS = ["/login", "/register", "/consent"];

export default function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { me } = useAuth();

  const bare =
    BARE_PATHS.includes(pathname) || (pathname === "/" && !me?.authenticated);

  if (bare) {
    return <div className="min-h-screen w-full">{children}</div>;
  }

  return (
    <>
      <AppSidebar />
      <SidebarInset className="flex min-h-svh flex-col bg-background">
        <Header />
        <div className="flex-1">{children}</div>
        <Footer />
      </SidebarInset>
    </>
  );
}
