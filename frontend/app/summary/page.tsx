"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  HeartPulse,
  Upload,
  Activity,
  Stethoscope,
  Pill,
  FlaskConical,
  ShieldCheck,
  RefreshCw,
} from "lucide-react";
import { api, type SummaryResponse, type SummaryIndicator, type SummaryMeasurement } from "@/lib/api";
import { cn } from "@/lib/utils";
import { AiChat } from "@/components/custom/AiChat";
import { AiOverviewCard } from "@/components/custom/AiOverviewCard";

function DeviationChip({ d }: { d: SummaryIndicator["deviation"] }) {
  const map: Record<string, { label: string; cls: string }> = {
    above: { label: "Выше нормы", cls: "border-warning/25 bg-warning/10 text-warning" },
    below: { label: "Ниже нормы", cls: "border-warning/25 bg-warning/10 text-warning" },
    within: { label: "В норме", cls: "border-success/25 bg-success/10 text-success" },
    unknown: {
      label: "Референс не указан",
      cls: "border-border bg-muted text-muted-foreground",
    },
  };
  const s = map[d] ?? map.unknown;
  return <span className={cn("status-chip", s.cls)}>{s.label}</span>;
}

const KIND_LABEL: Record<string, string> = {
  pressure: "Давление",
  pulse: "Пульс",
  weight: "Вес",
};

function fmtMeasurement(m: SummaryMeasurement): string {
  if (m.kind === "pressure")
    return `${m.systolic ?? "—"}/${m.diastolic ?? "—"} мм рт. ст.`;
  if (m.kind === "pulse") return `${m.pulse ?? "—"} уд/мин`;
  if (m.kind === "weight") return `${m.weight ?? "—"} кг`;
  return "—";
}

export default function SummaryPage() {
  const [data, setData] = useState<SummaryResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const load = () => {
    setLoading(true);
    api
      .get<SummaryResponse>("/api/summary")
      .then(setData)
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

  const abnormal = data?.abnormal ?? [];

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      {/* Заголовок */}
      <div className="mb-8 flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="mb-1 text-sm font-medium text-primary">Сводка</p>
          <h1 className="font-display text-3xl font-semibold text-foreground">
            Ваша медицинская сводка
          </h1>
          <p className="mt-2 max-w-2xl text-muted-foreground">
            Подтверждённые показатели, измерения и история в одном месте. Здесь же
            можно спросить ИИ-ассистента о своих данных.
          </p>
        </div>
        <button
          type="button"
          onClick={load}
          disabled={loading}
          className="flex cursor-pointer items-center justify-center gap-2 rounded-lg border border-border bg-background px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:border-primary/30 hover:bg-primary/5 disabled:cursor-default disabled:opacity-50"
        >
          <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
          Обновить
        </button>
      </div>

      {/* ИИ-обзор */}
      <div className="mb-6">
        <AiOverviewCard />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Левая колонка — данные */}
        <div className="space-y-6 lg:col-span-2">
          {/* Отклонения от референса (FR8) */}
          <div className="surface overflow-hidden">
            <div className="flex items-center gap-2.5 border-b border-border px-5 py-4">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-warning/10">
                <HeartPulse className="h-4 w-4 text-warning" />
              </span>
              <div>
                <h2 className="font-display font-semibold text-foreground">
                  Отклонения от референса
                </h2>
                <p className="text-xs text-muted-foreground">
                  {loading
                    ? "Загрузка…"
                    : abnormal.length > 0
                      ? `${abnormal.length} показател(я) вне референса`
                      : "Все подтверждённые показатели в пределах референса"}
                </p>
              </div>
            </div>
            <div className="divide-y divide-border">
              {loading && (
                <div className="p-5 text-sm text-muted-foreground">Загрузка…</div>
              )}
              {!loading && abnormal.length === 0 && (
                <div className="flex items-center gap-3 px-5 py-6 text-sm text-muted-foreground">
                  <ShieldCheck className="h-4 w-4 text-success" />
                  Отклонений среди подтверждённых показателей нет.
                </div>
              )}
              {abnormal.map((ind) => (
                <div key={ind.id} className="flex items-center gap-4 px-5 py-3.5">
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-foreground">
                      {ind.name}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Референс: {ind.reference_range || "не указан"}
                      {ind.test_date ? ` · ${ind.test_date}` : ""}
                    </p>
                  </div>
                  <span className="text-sm font-semibold text-foreground">
                    {ind.value} {ind.unit}
                  </span>
                  <DeviationChip d={ind.deviation} />
                </div>
              ))}
            </div>
          </div>

          {/* Все подтверждённые показатели */}
          <div className="surface overflow-hidden">
            <div className="flex items-center gap-2.5 border-b border-border px-5 py-4">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                <FlaskConical className="h-4 w-4 text-primary" />
              </span>
              <div>
                <h2 className="font-display font-semibold text-foreground">
                  Показатели анализов
                </h2>
                <p className="text-xs text-muted-foreground">
                  Только подтверждённые вами значения
                </p>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left text-xs uppercase tracking-wider text-muted-foreground">
                    <th className="px-5 py-3 font-medium">Показатель</th>
                    <th className="px-5 py-3 font-medium">Значение</th>
                    <th className="px-5 py-3 font-medium">Референс</th>
                    <th className="px-5 py-3 font-medium">Статус</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {loading && (
                    <tr>
                      <td colSpan={4} className="px-5 py-5 text-muted-foreground">
                        Загрузка…
                      </td>
                    </tr>
                  )}
                  {!loading && (data?.indicators.length ?? 0) === 0 && (
                    <tr>
                      <td colSpan={4} className="px-5 py-8 text-center text-muted-foreground">
                        Нет подтверждённых показателей. Загрузите бланк анализов
                        и подтвердите значения.
                      </td>
                    </tr>
                  )}
                  {data?.indicators.map((ind) => (
                    <tr key={ind.id}>
                      <td className="px-5 py-3">
                        <p className="font-medium text-foreground">{ind.name}</p>
                        {ind.test_date && (
                          <p className="text-xs text-muted-foreground">{ind.test_date}</p>
                        )}
                      </td>
                      <td className="px-5 py-3 font-semibold text-foreground">
                        {ind.value} <span className="text-xs font-normal text-muted-foreground">{ind.unit}</span>
                      </td>
                      <td className="px-5 py-3 text-muted-foreground">
                        {ind.reference_range || "—"}
                      </td>
                      <td className="px-5 py-3">
                        <DeviationChip d={ind.deviation} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Измерения + Диагнозы/Лекарства */}
          <div className="grid gap-6 sm:grid-cols-2">
            <div className="surface overflow-hidden">
              <div className="flex items-center gap-2.5 border-b border-border px-5 py-4">
                <Activity className="h-4 w-4 text-success" />
                <h2 className="font-display font-semibold text-foreground">
                  Измерения
                </h2>
              </div>
              <div className="divide-y divide-border">
                {loading && (
                  <div className="px-5 py-4 text-sm text-muted-foreground">Загрузка…</div>
                )}
                {!loading && (data?.measurements.length ?? 0) === 0 && (
                  <div className="px-5 py-6 text-sm text-muted-foreground">
                    Измерений пока нет.
                  </div>
                )}
                {data?.measurements.slice(0, 6).map((m, i) => (
                  <div key={i} className="flex items-center justify-between px-5 py-3">
                    <div>
                      <p className="text-sm font-medium text-foreground">
                        {KIND_LABEL[m.kind] ?? m.kind}
                      </p>
                      <p className="text-xs text-muted-foreground">{m.measured_at}</p>
                    </div>
                    <span className="text-sm font-semibold text-foreground">
                      {fmtMeasurement(m)}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className="surface overflow-hidden">
              <div className="flex items-center gap-2.5 border-b border-border px-5 py-4">
                <Stethoscope className="h-4 w-4 text-primary" />
                <h2 className="font-display font-semibold text-foreground">
                  Диагнозы и лекарства
                </h2>
              </div>
              <div className="divide-y divide-border">
                {loading && (
                  <div className="px-5 py-4 text-sm text-muted-foreground">Загрузка…</div>
                )}
                {!loading &&
                  (data?.diagnoses.length ?? 0) === 0 &&
                  (data?.medications.length ?? 0) === 0 && (
                    <div className="px-5 py-6 text-sm text-muted-foreground">
                      Записей пока нет.
                    </div>
                  )}
                {data?.diagnoses.map((d, i) => (
                  <div key={`d${i}`} className="flex items-center gap-3 px-5 py-3">
                    <Stethoscope className="h-4 w-4 flex-shrink-0 text-primary" />
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-foreground">
                        {d.name}
                      </p>
                      {d.comment && (
                        <p className="truncate text-xs text-muted-foreground">
                          {d.comment}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
                {data?.medications.map((m, i) => (
                  <div key={`m${i}`} className="flex items-center gap-3 px-5 py-3">
                    <Pill className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium text-foreground">
                        {m.name}
                      </p>
                      <p className="truncate text-xs text-muted-foreground">
                        {[m.drug_class, m.frequency].filter(Boolean).join(" · ") ||
                          "—"}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Пустое состояние */}
          {!loading && data?.empty && (
            <div className="surface flex flex-col items-center gap-4 px-6 py-10 text-center">
              <span className="flex h-12 w-12 items-center justify-center rounded-full bg-muted">
                <HeartPulse className="h-5 w-5 text-muted-foreground" />
              </span>
              <div>
                <p className="font-medium text-foreground">Сводка пока пуста</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Загрузите анализы или заполните профиль здоровья, чтобы собрать
                  сводку.
                </p>
              </div>
              <div className="flex flex-wrap justify-center gap-3">
                <Link
                  href="/analysis"
                  className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
                >
                  <Upload className="h-4 w-4" /> Загрузить анализы
                </Link>
                <Link
                  href="/profile"
                  className="flex items-center gap-2 rounded-lg border border-border bg-background px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:border-primary/30 hover:bg-primary/5"
                >
                  <Stethoscope className="h-4 w-4" /> Заполнить профиль
                </Link>
              </div>
            </div>
          )}
        </div>

        {/* Правая колонка — ИИ-чат */}
        <div className="lg:col-span-1">
          <AiChat className="h-[560px] lg:sticky lg:top-6" />
        </div>
      </div>
    </div>
  );
}
