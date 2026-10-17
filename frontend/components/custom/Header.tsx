"use client";

import { useEffect, useState } from "react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { ShieldCheck } from "lucide-react";
import ThemeToggle from "@/components/custom/ThemeToggle";

type Me = {
  authenticated: boolean;
  email?: string;
};

const Header: React.FC = () => {
  const [me, setMe] = useState<Me>({ authenticated: false });

  useEffect(() => {
    fetch("/api/auth/me", { credentials: "include" })
      .then((r) => r.json())
      .then((data) => setMe(data))
      .catch(() => setMe({ authenticated: false }));
  }, []);

  const initials = me.email ? me.email.slice(0, 2).toUpperCase() : "??";

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between gap-4 border-b border-border bg-background/85 px-4 backdrop-blur-md sm:px-6">
      <div className="flex items-center gap-3">
        <SidebarTrigger className="md:hidden" />
        <div className="hidden items-center gap-2 rounded-full border border-primary/15 bg-primary/5 px-3 py-1.5 sm:flex">
          <ShieldCheck className="h-4 w-4 text-primary" />
          <span className="text-xs font-medium text-primary">
            Информационная сводка · не медицинское заключение
          </span>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <ThemeToggle />
        <div className="flex items-center gap-3 rounded-xl border border-border bg-card py-1.5 pl-1.5 pr-3">
          <Avatar className="h-8 w-8">
            <AvatarFallback className="bg-primary/10 text-xs font-semibold text-primary">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="hidden max-w-[180px] leading-tight sm:block">
            <p className="truncate text-sm font-medium text-foreground">
              {me.email ?? "Гость"}
            </p>
            <p className="text-[0.7rem] text-muted-foreground">Пациент</p>
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
