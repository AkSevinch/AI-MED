"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ShieldCheck } from "lucide-react";
import { api } from "@/lib/api";
import { useAuth } from "@/components/custom/AuthGate";
import { Button } from "@/components/ui/button";

const CONSENT_POINTS = [
  "AI MED собирает и обрабатывает медицинские данные, которые вы загружаете или вводите сами.",
  "Сервис формирует информационную сводку и не является медицинским заключением, диагнозом или назначением.",
  "Сервис не анализирует снимки (ЭКГ, УЗИ, КТ, МРТ, рентген) и не даёт рекомендаций по лечению или отмене препаратов.",
  "Значения, распознанные автоматически, попадают в сводку только после вашего подтверждения.",
  "Вы можете в любой момент удалить свои данные и аккаунт.",
];

export default function ConsentPage() {
  const router = useRouter();
  const { me, refresh } = useAuth();
  const [agreed, setAgreed] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const accept = async () => {
    if (!agreed) return;
    setError("");
    setBusy(true);
    try {
      await api.post("/api/consent", { version: "1.0" });
      await refresh();
      router.replace("/");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  };

  const logout = async () => {
    try {
      await api.post("/api/auth/logout");
    } catch {}
    await refresh();
    router.replace("/login");
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6 py-10">
      <div className="w-full max-w-2xl">
        <div className="surface overflow-hidden">
          <div className="flex items-center gap-3 border-b border-border bg-primary/5 px-7 py-5">
            <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10">
              <ShieldCheck className="h-5 w-5 text-primary" />
            </span>
            <div>
              <h1 className="font-display text-xl font-semibold text-foreground">
                Согласие на обработку данных
              </h1>
              <p className="text-sm text-muted-foreground">
                {me?.email ? `Аккаунт: ${me.email}` : "Перед началом работы"}
              </p>
            </div>
          </div>

          <div className="p-7">
            <p className="mb-5 text-sm text-muted-foreground">
              Чтобы начать пользоваться сервисом, ознакомьтесь с условиями и
              подтвердите согласие.
            </p>

            <ul className="mb-6 space-y-3">
              {CONSENT_POINTS.map((point, i) => (
                <li key={i} className="flex items-start gap-3">
                  <span className="mt-1.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-primary" />
                  <span className="text-sm leading-relaxed text-foreground">
                    {point}
                  </span>
                </li>
              ))}
            </ul>

            <label className="mb-6 flex cursor-pointer items-start gap-3 rounded-xl border border-primary/20 bg-primary/5 p-4">
              <input
                type="checkbox"
                checked={agreed}
                onChange={(e) => setAgreed(e.target.checked)}
                className="mt-0.5 h-4 w-4 cursor-pointer accent-primary"
              />
              <span className="text-sm text-foreground">
                Я прочитал(а) и принимаю условия обработки медицинских данных
                (версия 1.0).
              </span>
            </label>

            {error && (
              <p className="mb-4 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                {error}
              </p>
            )}

            <div className="flex items-center justify-between gap-4">
              <button
                type="button"
                onClick={logout}
                className="cursor-pointer text-sm text-muted-foreground transition-colors hover:text-foreground"
              >
                Выйти
              </button>
              <Button
                onClick={accept}
                size="lg"
                disabled={!agreed || busy}
                className="px-8"
              >
                {busy ? "Сохранение…" : "Принять и продолжить"}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
