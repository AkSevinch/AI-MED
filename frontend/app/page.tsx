"use client";

import { useEffect, useState, type ComponentType } from "react";
import Link from "next/link";
import {
  FlaskConical,
  Stethoscope,
  Activity,
  ClipboardList,
  FileText,
  ArrowRight,
  Upload,
  ShieldCheck,
  Clock,
  FilePlus2,
} from "lucide-react";
import { api } from "@/lib/api";
import { useAuth } from "@/components/custom/AuthGate";
import { Button } from "@/components/ui/button";
import { DocStatusChip } from "@/components/custom/StatusChip";
import { AiOverviewCard } from "@/components/custom/AiOverviewCard";
import Landing from "@/components/custom/Landing";

type Doc = {
  id: number;
  filename: string;
  status: string;
  uploaded_at: string;
  test_date: string | null;
};

type Counts = {
  documents: number;
  pending: number;
  diagnoses: number;
  measurements: number;
};

const EMPTY: Counts = { documents: 0, pending: 0, diagnoses: 0, measurements: 0 };

function StatCard({
  icon: Icon,
  label,
  value,
  tone = "primary",
}: {
  icon: ComponentType<{ className?: string }>;
  label: string;
  value: number | string;
  tone?: "primary" | "warning" | "success";
}) {
  const toneCls =
    tone === "warning"
      ? "bg-warning/10 text-warning"
      : tone === "success"
        ? "bg-success/10 text-success"
        : "bg-primary/10 text-primary";
  return (
    <div className="surface p-5">
      <div className={`mb-3 flex h-9 w-9 items-center justify-center rounded-lg ${toneCls}`}>
        <Icon className="h-[18px] w-[18px]" />
      </div>
      <p className="font-display text-2xl font-semibold text-foreground">{value}</p>
      <p className="text-sm text-muted-foreground">{label}</p>
    </div>
  );
}

export default function OverviewPage() {
  const { me } = useAuth();
  const [counts, setCounts] = useState<Counts>(EMPTY);
  const [docs, setDocs] = useState<Doc[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    (async () => {
      const safe = <T,>(p: Promise<T>, fallback: T) => p.catch(() => fallback);
      const [documents, diagnoses, measurements] = await Promise.all([
        safe(api.get<Doc[]>("/api/documents"), []),
        safe(api.get<unknown[]>("/api/diagnoses"), []),
        safe(api.get<unknown[]>("/api/measurements"), []),
      ]);
      if (!alive) return;
      setDocs(documents);
      setCounts({
        documents: documents.length,
        pending: documents.filter((d) => d.status === "needs_review").length,
        diagnoses: diagnoses.length,
        measurements: measurements.length,
      });
      setLoading(false);
    })();
    return () => {
      alive = false;
    };
  }, []);

  const recent = docs.slice(0, 5);

  if (!me?.authenticated) {
    return <Landing />;
  }

  return (
    <div className="mx-auto max-w-6xl px-6 py-8">
      {/* Заголовок */}
      <div className="mb-8 flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="mb-1 text-sm font-medium text-primary">Обзор</p>
          <h1 className="font-display text-3xl font-semibold text-foreground">
            Ваша сводка здоровья
          </h1>
          <p className="mt-2 max-w-2xl text-muted-foreground">
            {me?.email ? `${me.email} · ` : ""}
            Соберите анализы, измерения и историю в одну понятную картину. В
            сводку попадают только подтверждённые вами значения.
          </p>
        </div>
        <Button asChild size="lg">
          <Link href="/analysis">
            <Upload className="h-4 w-4" />
            Загрузить анализы
          </Link>
        </Button>
      </div>

      {/* Статистика */}
      <div className="mb-8 grid grid-cols-2 gap-4 lg:grid-cols-4">
        {loading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="surface h-[124px] animate-pulse bg-muted/40" />
          ))
        ) : (
          <>
            <StatCard icon={FileText} label="Документы" value={counts.documents} />
            <StatCard
              icon={Clock}
              label="Ожидают проверки"
              value={counts.pending}
              tone="warning"
            />
            <StatCard
              icon={Stethoscope}
              label="Диагнозы"
              value={counts.diagnoses}
            />
            <StatCard
              icon={Activity}
              label="Измерения"
              value={counts.measurements}
              tone="success"
            />
          </>
        )}
      </div>

      {/* ИИ-обзор подтверждённых данных */}
      <div className="mb-6">
        <AiOverviewCard />
      </div>

      <div className="mb-8 grid gap-6 lg:grid-cols-3">
        {/* Последние документы */}
        <div className="surface lg:col-span-2">
          <div className="flex items-center justify-between border-b border-border px-5 py-4">
            <h2 className="font-display font-semibold text-foreground">
              Последние документы
            </h2>
            <Link
              href="/analysis"
              className="flex items-center gap-1 text-sm font-medium text-primary hover:underline"
            >
              Все анализы <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>
          <div className="divide-y divide-border">
            {loading && (
              <div className="p-5 text-sm text-muted-foreground">Загрузка…</div>
            )}
            {!loading && recent.length === 0 && (
              <div className="flex flex-col items-center gap-3 px-5 py-12 text-center">
                <span className="flex h-11 w-11 items-center justify-center rounded-full bg-muted">
                  <FlaskConical className="h-5 w-5 text-muted-foreground" />
                </span>
                <p className="text-sm text-muted-foreground">
                  Пока нет загруженных документов
                </p>
                <Button asChild variant="outline" size="sm">
                  <Link href="/analysis">
                    <FilePlus2 className="h-4 w-4" /> Загрузить первый бланк
                  </Link>
                </Button>
              </div>
            )}
            {recent.map((d) => (
              <Link
                key={d.id}
                href="/analysis"
                className="flex items-center gap-4 px-5 py-3.5 transition-colors hover:bg-muted/40"
              >
                <span className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10">
                  <FileText className="h-4 w-4 text-primary" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-foreground">
                    {d.filename}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {new Date(d.uploaded_at).toLocaleDateString("ru-RU")}
                    {d.test_date ? ` · дата анализа ${d.test_date}` : ""}
                  </p>
                </div>
                <DocStatusChip status={d.status} />
              </Link>
            ))}
          </div>
        </div>

        {/* Боковая колонка */}
        <div className="space-y-6">
          <div className="surface p-5">
            <h2 className="mb-4 font-display font-semibold text-foreground">
              Дополнить профиль
            </h2>
            <div className="space-y-2">
              <Link
                href="/profile"
                className="flex items-center gap-3 rounded-lg border border-border px-3.5 py-3 transition-colors hover:border-primary/30 hover:bg-primary/5"
              >
                <ClipboardList className="h-4 w-4 text-primary" />
                <span className="flex-1 text-sm font-medium text-foreground">
                  Заполнить анкету
                </span>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
              </Link>
              <Link
                href="/profile"
                className="flex items-center gap-3 rounded-lg border border-border px-3.5 py-3 transition-colors hover:border-primary/30 hover:bg-primary/5"
              >
                <Stethoscope className="h-4 w-4 text-primary" />
                <span className="flex-1 text-sm font-medium text-foreground">
                  Диагнозы и лекарства
                </span>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
              </Link>
              <Link
                href="/profile"
                className="flex items-center gap-3 rounded-lg border border-border px-3.5 py-3 transition-colors hover:border-primary/30 hover:bg-primary/5"
              >
                <Activity className="h-4 w-4 text-primary" />
                <span className="flex-1 text-sm font-medium text-foreground">
                  Добавить измерение
                </span>
                <ArrowRight className="h-4 w-4 text-muted-foreground" />
              </Link>
            </div>
          </div>

          <div className="surface overflow-hidden">
            <div className="flex items-center gap-2 bg-success/10 px-5 py-3">
              <ShieldCheck className="h-4 w-4 text-success" />
              <p className="text-sm font-semibold text-success">
                Как мы формируем сводку
              </p>
            </div>
            <ul className="space-y-3 p-5 text-sm text-muted-foreground">
              <li className="flex gap-2.5">
                <span className="mt-1.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-success" />
                Загружаете бланк — распознаём показатели.
              </li>
              <li className="flex gap-2.5">
                <span className="mt-1.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-success" />
                Вы проверяете и подтверждаете каждое значение.
              </li>
              <li className="flex gap-2.5">
                <span className="mt-1.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-success" />
                Только подтверждённые данные попадают в сводку.
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
