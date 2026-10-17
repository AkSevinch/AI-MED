"use client";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  LayoutDashboard,
  FlaskConical,
  HeartPulse,
  UserRound,
  ShieldCheck,
  LogOut,
} from "lucide-react";
import { useAuth } from "@/components/custom/AuthGate";

const NAV = [
  { href: "/", label: "Обзор", icon: LayoutDashboard },
  { href: "/summary", label: "Сводка", icon: HeartPulse },
  { href: "/analysis", label: "Анализы", icon: FlaskConical },
  { href: "/profile", label: "Профиль здоровья", icon: UserRound },
];

function BrandMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" className={className} aria-hidden="true">
      <rect width="32" height="32" rx="9" fill="var(--primary)" />
      <path
        d="M9 16h3.2l1.9-5 3.4 10 2-6.2 1.4 1.2H23"
        fill="none"
        stroke="#fff"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function AppSidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const { me } = useAuth();

  const handleLogout = async () => {
    try {
      await fetch("/api/auth/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch {}
    router.replace("/login");
  };

  return (
    <Sidebar className="border-r border-sidebar-border">
      <SidebarHeader className="p-3">
        <Link
          href="/"
          className="flex items-center gap-3 rounded-xl px-2 py-2 transition-colors hover:bg-sidebar-accent/60"
        >
          <BrandMark className="h-9 w-9 flex-shrink-0" />
          <div className="leading-tight">
            <p className="font-display font-semibold text-[0.95rem] text-sidebar-foreground">
              AI MED
            </p>
            <p className="text-xs text-muted-foreground">
              Сводка здоровья
            </p>
          </div>
        </Link>
      </SidebarHeader>

      <SidebarContent className="px-2">
        <SidebarGroup>
          <SidebarGroupLabel className="text-[0.7rem] font-semibold uppercase tracking-wider text-muted-foreground">
            Навигация
          </SidebarGroupLabel>
          <SidebarMenu>
            {NAV.map((item) => {
              const Icon = item.icon;
              const active = pathname === item.href;
              return (
                <SidebarMenuItem key={item.href}>
                  <SidebarMenuButton
                    asChild
                    isActive={active}
                    className={
                      active
                        ? "relative bg-sidebar-accent font-semibold text-sidebar-accent-foreground before:absolute before:left-0 before:top-1/2 before:h-5 before:w-1 before:-translate-y-1/2 before:rounded-r-full before:bg-primary"
                        : "text-sidebar-foreground hover:bg-sidebar-accent/60"
                    }
                  >
                    <Link href={item.href}>
                      <Icon className="h-[1.15rem] w-[1.15rem]" />
                      <span>{item.label}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })}
          </SidebarMenu>
        </SidebarGroup>

        <SidebarGroup className="mt-2">
          <div className="rounded-xl border border-primary/15 bg-primary/5 p-3.5">
            <div className="mb-1.5 flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-primary" />
              <p className="text-sm font-semibold text-primary">
                Безопасность
              </p>
            </div>
            <p className="text-xs leading-relaxed text-muted-foreground">
              В сводку попадают только значения, которые вы подтвердили вручную.
            </p>
          </div>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="p-3">
        <div className="rounded-xl border border-sidebar-border bg-sidebar-accent/40 p-3">
          <p className="truncate text-xs font-medium text-sidebar-foreground">
            {me?.email ?? "Гость"}
          </p>
          <p className="text-[0.7rem] text-muted-foreground">Пациент</p>
          <button
            type="button"
            onClick={handleLogout}
            className="mt-3 flex w-full cursor-pointer items-center justify-center gap-2 rounded-lg border border-sidebar-border bg-background px-3 py-2 text-xs font-medium text-muted-foreground transition-colors duration-200 hover:border-destructive/30 hover:text-destructive"
          >
            <LogOut className="h-3.5 w-3.5" />
            Выйти
          </button>
        </div>
      </SidebarFooter>
    </Sidebar>
  );
}
