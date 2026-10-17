"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { useAuth } from "@/components/custom/AuthGate";
import AuthLayout from "@/components/custom/AuthLayout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const POINTS = [
  "Анализы, измерения и история — в одной понятной сводке.",
  "Каждое распознанное значение вы подтверждаете вручную.",
  "Данные привязаны к вашему аккаунту и видны только вам.",
];

export default function LoginPage() {
  const router = useRouter();
  const { refresh } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await api.post("/api/auth/login", { email, password });
      await refresh();
      router.replace("/");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthLayout
      title="Соберите историю здоровья в одну сводку"
      subtitle="Единая сводка здоровья"
      points={POINTS}
    >
      <div className="mb-8">
        <h1 className="font-display text-2xl font-semibold text-foreground">
          Вход
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Войдите, чтобы получить доступ к своим данным
        </p>
      </div>

      <form onSubmit={submit} className="space-y-5">
        <div className="space-y-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            autoComplete="email"
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">Пароль</Label>
          <Input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            autoComplete="current-password"
            required
          />
        </div>

        {error && (
          <p className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive">
            {error}
          </p>
        )}

        <Button type="submit" size="lg" disabled={busy} className="w-full">
          {busy ? "Вход…" : "Войти"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Нет аккаунта?{" "}
        <Link href="/register" className="font-medium text-primary hover:underline">
          Зарегистрироваться
        </Link>
      </p>
    </AuthLayout>
  );
}
