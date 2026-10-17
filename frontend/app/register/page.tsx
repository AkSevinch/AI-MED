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
  "Бесплатный аккаунт для ведения личной сводки здоровья.",
  "Загружайте бланки анализов в форматах PDF, фото или DOCX.",
  "Вы контролируете, какие значения попадут в сводку.",
];

export default function RegisterPage() {
  const router = useRouter();
  const { refresh } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (password.length < 8) {
      setError("Пароль должен быть не короче 8 символов");
      return;
    }
    if (password !== confirm) {
      setError("Пароли не совпадают");
      return;
    }
    setBusy(true);
    try {
      await api.post("/api/auth/register", { email, password });
      await refresh();
      router.replace("/consent");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthLayout
      title="Начните вести единую сводку здоровья"
      subtitle="Создание аккаунта"
      points={POINTS}
    >
      <div className="mb-8">
        <h1 className="font-display text-2xl font-semibold text-foreground">
          Регистрация
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Создайте аккаунт для ведения своей сводки здоровья
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
            placeholder="Минимум 8 символов"
            autoComplete="new-password"
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="confirm">Повторите пароль</Label>
          <Input
            id="confirm"
            type="password"
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            placeholder="••••••••"
            autoComplete="new-password"
            required
          />
        </div>

        {error && (
          <p className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive">
            {error}
          </p>
        )}

        <Button type="submit" size="lg" disabled={busy} className="w-full">
          {busy ? "Создание…" : "Создать аккаунт"}
        </Button>
      </form>

      <p className="mt-6 text-center text-sm text-muted-foreground">
        Уже есть аккаунт?{" "}
        <Link href="/login" className="font-medium text-primary hover:underline">
          Войти
        </Link>
      </p>
    </AuthLayout>
  );
}
