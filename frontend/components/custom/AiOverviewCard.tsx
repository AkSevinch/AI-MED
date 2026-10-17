"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { Sparkles, RefreshCw, AlertTriangle, ArrowRight, ShieldCheck } from "lucide-react";
import { api, type OverviewResponse } from "@/lib/api";

type State =
  | { status: "loading" }
  | { status: "ready"; text: string; empty: boolean }
  | { status: "error"; message: string };

function renderLine(line: string, i: number) {
  const clean = line.replace(/^#{1,3}\s*/, "").trim();
  const heading =
    /^(Кратко|Отклонения от референса|Что обсудить с врачом|Отклонения|Вопросы врачу)/i.test(
      clean,
    ) && clean.length < 60;
  if (heading) {
    return (
      <h3
        key={i}
        className="mt-3 font-display text-sm font-semibold text-foreground first:mt-0"
      >
        {clean.replace(/:$/, "")}
      </h3>
    );
  }
  const bullet = /^([-*•]|\d+[.)])\s+/.test(clean);
  if (bullet) {
    return (
      <p key={i} className="flex gap-2.5 text-sm text-muted-foreground">
        <span className="mt-1.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-primary" />
        <span>{clean.replace(/^([-*•]|\d+[.)])\s+/, "")}</span>
      </p>
    );
  }
  if (!clean) return <div key={i} className="h-1" />;
  return (
    <p key={i} className="text-sm leading-relaxed text-muted-foreground">
      {clean}
    </p>
  );
}

export function AiOverviewCard() {
  const [state, setState] = useState<State>({ status: "loading" });

  const load = useCallback(async () => {
    setState({ status: "loading" });
    try {
      const res = await api.post<OverviewResponse>("/api/summary/overview");
      setState({ status: "ready", text: res.text, empty: res.empty });
    } catch (e) {
      setState({
        status: "error",
        message: e instanceof Error ? e.message : "Не удалось получить обзор",
      });
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <div className="surface overflow-hidden">
      <div className="flex items-center justify-between border-b border-border px-5 py-4">
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
            <Sparkles className="h-4 w-4 text-primary" />
          </span>
          <div>
            <h2 className="font-display font-semibold text-foreground">
              ИИ-обзор
            </h2>
            <p className="text-xs text-muted-foreground">
              Коротко о ваших подтверждённых данных
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={load}
          disabled={state.status === "loading"}
          className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:cursor-default disabled:opacity-50"
          aria-label="Обновить обзор"
        >
          <RefreshCw
            className={`h-4 w-4 ${state.status === "loading" ? "animate-spin" : ""}`}
          />
        </button>
      </div>

      <div className="px-5 py-4">
        {state.status === "loading" && (
          <div className="space-y-2.5">
            <div className="h-3.5 w-3/4 animate-pulse rounded bg-muted" />
            <div className="h-3.5 w-full animate-pulse rounded bg-muted" />
            <div className="h-3.5 w-5/6 animate-pulse rounded bg-muted" />
          </div>
        )}

        {state.status === "error" && (
          <div className="flex flex-col gap-3">
            <div className="flex items-start gap-2.5 rounded-lg border border-warning/25 bg-warning/10 p-3">
              <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0 text-warning" />
              <div className="text-sm text-warning">
                <p className="font-medium">ИИ-обзор временно недоступен</p>
                <p className="mt-0.5 text-xs leading-relaxed text-warning/90">
                  {state.message}
                </p>
              </div>
            </div>
            <Link
              href="/summary"
              className="flex items-center gap-1 text-sm font-medium text-primary hover:underline"
            >
              Посмотреть сводку без ИИ <ArrowRight className="h-3.5 w-3.5" />
            </Link>
          </div>
        )}

        {state.status === "ready" && (
          <div className="space-y-1.5">
            {state.empty ? (
              <p className="text-sm leading-relaxed text-muted-foreground">
                {state.text}
              </p>
            ) : (
              state.text
                .split("\n")
                .map((line, i) => renderLine(line, i))
            )}
          </div>
        )}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-border bg-muted/30 px-5 py-3">
        <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <ShieldCheck className="h-3.5 w-3.5 text-success" />
          ИИ не ставит диагноз и не назначает лечение
        </p>
        <Link
          href="/summary"
          className="flex items-center gap-1 text-sm font-medium text-primary hover:underline"
        >
          Открыть сводку <ArrowRight className="h-3.5 w-3.5" />
        </Link>
      </div>
    </div>
  );
}
