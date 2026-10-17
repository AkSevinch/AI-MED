"use client";

import { useEffect, useState } from "react";
import { Moon, Sun } from "lucide-react";

const STORAGE_KEY = "aimed-theme";

export default function ThemeToggle({ className = "" }: { className?: string }) {
  const [dark, setDark] = useState<boolean | null>(null);

  useEffect(() => {
    setDark(document.documentElement.classList.contains("dark"));
  }, []);

  const toggle = () => {
    const next = !document.documentElement.classList.contains("dark");
    document.documentElement.classList.toggle("dark", next);
    try {
      localStorage.setItem(STORAGE_KEY, next ? "dark" : "light");
    } catch {
      // локальное хранилище недоступно — тема применится только на сессию
    }
    setDark(next);
  };

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={dark ? "Включить светлую тему" : "Включить тёмную тему"}
      title={dark ? "Светлая тема" : "Тёмная тема"}
      className={
        "inline-flex h-9 w-9 cursor-pointer items-center justify-center rounded-lg border border-border bg-card text-foreground transition-colors hover:bg-accent hover:text-accent-foreground " +
        className
      }
    >
      {dark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </button>
  );
}
