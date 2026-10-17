"use client";

import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  Upload,
  FileText,
  RefreshCw,
  Trash2,
  Check,
  X,
  ExternalLink,
  ShieldCheck,
  Loader2,
  Save,
  FlaskConical,
} from "lucide-react";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { DocStatusChip, VerificationChip } from "@/components/custom/StatusChip";

type Doc = {
  id: number;
  filename: string;
  mime: string;
  status: string;
  lab_name: string;
  test_date: string;
  uploaded_at: string;
};

type Indicator = {
  id: number;
  code: string;
  name: string;
  value: string;
  unit: string;
  reference_range: string;
  test_date: string | null;
  source: string;
  verification_status: string;
  confidence: number | null;
};

const ALLOWED = ".pdf,.jpg,.jpeg,.png,.docx";

function CellInput(props: React.ComponentProps<"input">) {
  return (
    <input
      {...props}
      className={`h-8 w-full rounded-md border border-transparent bg-transparent px-2 text-sm text-foreground transition-colors placeholder:text-muted-foreground/60 hover:border-border focus:border-ring focus:bg-background focus:outline-none focus:ring-1 focus:ring-ring/40 ${props.className ?? ""}`}
    />
  );
}

export default function AnalysisPage() {
  const [docs, setDocs] = useState<Doc[]>([]);
  const [selected, setSelected] = useState<Doc | null>(null);
  const [indicators, setIndicators] = useState<Indicator[]>([]);
  const [uploading, setUploading] = useState(false);
  const [loadingIndicators, setLoadingIndicators] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [message, setMessage] = useState<{ kind: "error" | "info"; text: string } | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const loadDocs = useCallback(async () => {
    try {
      setDocs(await api.get<Doc[]>("/api/documents"));
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    }
  }, []);

  useEffect(() => {
    loadDocs();
  }, [loadDocs]);

  const loadIndicators = useCallback(async (doc: Doc) => {
    setLoadingIndicators(true);
    try {
      setIndicators(await api.get<Indicator[]>(`/api/documents/${doc.id}/indicators`));
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    } finally {
      setLoadingIndicators(false);
    }
  }, []);

  const selectDoc = useCallback(
    (doc: Doc) => {
      setSelected(doc);
      setMessage(null);
      loadIndicators(doc);
    },
    [loadIndicators],
  );

  const upload = useCallback(
    async (file: File) => {
      setMessage(null);
      setUploading(true);
      try {
        const form = new FormData();
        form.append("file", file);
        const res = await api.post<{ document: Doc; indicators: Indicator[]; warning?: string }>(
          "/api/documents",
          form,
        );
        if (res.warning) setMessage({ kind: "info", text: res.warning });
        await loadDocs();
        setSelected(res.document);
        setIndicators(res.indicators ?? []);
      } catch (e) {
        setMessage({ kind: "error", text: (e as Error).message });
      } finally {
        setUploading(false);
        if (fileRef.current) fileRef.current.value = "";
      }
    },
    [loadDocs],
  );

  const removeDoc = async (id: number) => {
    try {
      await api.del(`/api/documents/${id}`);
      if (selected?.id === id) {
        setSelected(null);
        setIndicators([]);
      }
      await loadDocs();
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    }
  };

  const updateIndicatorLocal = (id: number, patch: Partial<Indicator>) =>
    setIndicators((prev) => prev.map((i) => (i.id === id ? { ...i, ...patch } : i)));

  const saveIndicator = async (ind: Indicator) => {
    try {
      await api.put(`/api/indicators/${ind.id}`, {
        name: ind.name,
        value: ind.value,
        unit: ind.unit,
        reference_range: ind.reference_range,
        test_date: ind.test_date ?? "",
      });
      setMessage({ kind: "info", text: "Показатель сохранён" });
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    }
  };

  const setStatus = async (ind: Indicator, status: "confirmed" | "needs_review") => {
    try {
      await api.post(
        `/api/indicators/${ind.id}/${status === "confirmed" ? "confirm" : "needs-review"}`,
      );
      updateIndicatorLocal(ind.id, { verification_status: status });
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    }
  };

  const removeIndicator = async (id: number) => {
    try {
      await api.del(`/api/indicators/${id}`);
      setIndicators((prev) => prev.filter((i) => i.id !== id));
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    }
  };

  const confirmAll = async () => {
    if (!selected) return;
    try {
      await api.post(`/api/documents/${selected.id}/confirm-all`);
      setIndicators((prev) => prev.map((i) => ({ ...i, verification_status: "confirmed" })));
      setSelected({ ...selected, status: "saved" });
      await loadDocs();
    } catch (e) {
      setMessage({ kind: "error", text: (e as Error).message });
    }
  };

  const confirmedCount = useMemo(
    () => indicators.filter((i) => i.verification_status === "confirmed").length,
    [indicators],
  );
  const progress = indicators.length
    ? Math.round((confirmedCount / indicators.length) * 100)
    : 0;

  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Заголовок */}
      <div className="mb-8">
        <p className="mb-1 text-sm font-medium text-primary">Работа с данными</p>
        <h1 className="font-display text-3xl font-semibold text-foreground">
          Анализы
        </h1>
        <p className="mt-2 max-w-3xl text-muted-foreground">
          Загрузите бланк анализов (PDF, JPG, PNG, DOCX). Мы распознаём
          показатели, а вы подтверждаете каждое значение перед добавлением в
          сводку.
        </p>
      </div>

      {message && (
        <div
          role="status"
          className={`mb-6 flex items-start justify-between gap-4 rounded-xl border px-4 py-3 text-sm ${
            message.kind === "error"
              ? "border-destructive/20 bg-destructive/5 text-destructive"
              : "border-primary/20 bg-primary/5 text-primary"
          }`}
        >
          <span>{message.text}</span>
          <button
            onClick={() => setMessage(null)}
            className="cursor-pointer opacity-60 transition-opacity hover:opacity-100"
            aria-label="Закрыть"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )}

      {/* Загрузка */}
      <div className="surface mb-8 p-1.5">
        <label
          onDragOver={(e) => {
            e.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={(e) => {
            e.preventDefault();
            setDragging(false);
            const f = e.dataTransfer.files?.[0];
            if (f) upload(f);
          }}
          className={`flex cursor-pointer flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed px-6 py-10 text-center transition-colors ${
            dragging
              ? "border-primary bg-primary/5"
              : uploading
                ? "border-primary/40 bg-primary/5"
                : "border-border hover:border-primary/50 hover:bg-muted/40"
          }`}
        >
          <input
            ref={fileRef}
            type="file"
            accept={ALLOWED}
            className="hidden"
            disabled={uploading}
            onChange={(e) => {
              const f = e.target.files?.[0];
              if (f) upload(f);
            }}
          />
          {uploading ? (
            <>
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <span className="text-sm font-medium text-primary">
                Распознавание документа…
              </span>
            </>
          ) : (
            <>
              <span className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                <Upload className="h-5 w-5 text-primary" />
              </span>
              <div>
                <span className="text-sm font-medium text-foreground">
                  Нажмите или перетащите файл
                </span>
                <p className="mt-1 text-xs text-muted-foreground">
                  PDF, JPG, JPEG, PNG, DOCX · до 20 МБ
                </p>
              </div>
            </>
          )}
        </label>
      </div>

      <div className="grid gap-6 lg:grid-cols-5">
        {/* Документы */}
        <div className="lg:col-span-2">
          <div className="surface overflow-hidden">
            <div className="flex items-center justify-between border-b border-border px-5 py-4">
              <h2 className="font-display font-semibold text-foreground">
                Мои документы
              </h2>
              <button
                onClick={loadDocs}
                className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                aria-label="Обновить список"
              >
                <RefreshCw className="h-4 w-4" />
              </button>
            </div>
            <div className="space-y-2 p-3">
              {docs.length === 0 && (
                <p className="px-3 py-6 text-center text-sm text-muted-foreground">
                  Документы ещё не загружены
                </p>
              )}
              {docs.map((d) => {
                const active = selected?.id === d.id;
                return (
                  <div
                    key={d.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => selectDoc(d)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") selectDoc(d);
                    }}
                    className={`cursor-pointer rounded-xl border p-3.5 transition-colors ${
                      active
                        ? "border-primary/40 bg-primary/5"
                        : "border-border hover:border-primary/25 hover:bg-muted/40"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex min-w-0 items-start gap-3">
                        <FileText className="mt-0.5 h-5 w-5 flex-shrink-0 text-primary" />
                        <div className="min-w-0">
                          <p className="truncate text-sm font-medium text-foreground">
                            {d.filename}
                          </p>
                          <p className="mt-0.5 text-xs text-muted-foreground">
                            {new Date(d.uploaded_at).toLocaleDateString("ru-RU")}
                            {d.test_date ? ` · ${d.test_date}` : ""}
                          </p>
                        </div>
                      </div>
                      <DocStatusChip status={d.status} />
                    </div>
                    <div className="mt-3 flex items-center gap-4">
                      <a
                        href={`/api/documents/${d.id}/file`}
                        target="_blank"
                        rel="noreferrer"
                        onClick={(e) => e.stopPropagation()}
                        className="flex items-center gap-1 text-xs font-medium text-primary hover:underline"
                      >
                        <ExternalLink className="h-3 w-3" /> Открыть
                      </a>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          removeDoc(d.id);
                        }}
                        className="flex cursor-pointer items-center gap-1 text-xs text-muted-foreground transition-colors hover:text-destructive"
                      >
                        <Trash2 className="h-3 w-3" /> Удалить
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Проверка */}
        <div className="lg:col-span-3">
          <div className="surface flex h-full flex-col overflow-hidden">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-5 py-4">
              <div className="min-w-0">
                <h2 className="font-display font-semibold text-foreground">
                  Проверка показателей
                </h2>
                <p className="truncate text-xs text-muted-foreground">
                  {selected ? selected.filename : "Выберите документ слева"}
                </p>
              </div>
              {indicators.length > 0 && (
                <div className="flex items-center gap-3">
                  <span className="hidden text-xs text-muted-foreground sm:inline">
                    Подтверждено {confirmedCount} из {indicators.length}
                  </span>
                  <Button size="sm" variant="cta" onClick={confirmAll}>
                    <Check className="h-4 w-4" /> Подтвердить все
                  </Button>
                </div>
              )}
            </div>

            {indicators.length > 0 && (
              <div className="h-1 w-full bg-muted">
                <div
                  className="h-full bg-success transition-all duration-300"
                  style={{ width: `${progress}%` }}
                />
              </div>
            )}

            <div className="flex-1">
              {!selected && (
                <div className="flex flex-col items-center py-20 text-center text-muted-foreground">
                  <FlaskConical className="mb-3 h-8 w-8" />
                  <p className="text-sm">
                    Выберите документ, чтобы увидеть распознанные показатели
                  </p>
                </div>
              )}

              {selected && loadingIndicators && (
                <div className="flex justify-center py-20">
                  <Loader2 className="h-6 w-6 animate-spin text-primary" />
                </div>
              )}

              {selected && !loadingIndicators && indicators.length === 0 && (
                <div className="flex flex-col items-center py-20 text-center text-muted-foreground">
                  <FlaskConical className="mb-3 h-8 w-8" />
                  <p className="text-sm">
                    {selected.status === "failed"
                      ? "Не удалось распознать показатели в этом документе"
                      : "Показатели не найдены"}
                  </p>
                </div>
              )}

              {selected && !loadingIndicators && indicators.length > 0 && (
                <Table>
                  <TableHeader className="bg-muted/40">
                    <TableRow className="hover:bg-transparent">
                      <TableHead className="min-w-[150px]">Показатель</TableHead>
                      <TableHead className="w-[90px]">Значение</TableHead>
                      <TableHead className="w-[80px]">Ед.</TableHead>
                      <TableHead className="w-[110px]">Референс</TableHead>
                      <TableHead className="w-[130px]">Дата</TableHead>
                      <TableHead className="w-[140px]">Статус</TableHead>
                      <TableHead className="w-[110px] text-right">Действия</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {indicators.map((ind) => {
                      const confirmed = ind.verification_status === "confirmed";
                      return (
                        <TableRow key={ind.id}>
                          <TableCell className="p-1">
                            <CellInput
                              value={ind.name}
                              aria-label="Название показателя"
                              onChange={(e) =>
                                updateIndicatorLocal(ind.id, { name: e.target.value })
                              }
                            />
                          </TableCell>
                          <TableCell className="p-1">
                            <CellInput
                              value={ind.value}
                              aria-label="Значение"
                              className="font-medium"
                              onChange={(e) =>
                                updateIndicatorLocal(ind.id, { value: e.target.value })
                              }
                            />
                          </TableCell>
                          <TableCell className="p-1">
                            <CellInput
                              value={ind.unit}
                              aria-label="Единица измерения"
                              onChange={(e) =>
                                updateIndicatorLocal(ind.id, { unit: e.target.value })
                              }
                            />
                          </TableCell>
                          <TableCell className="p-1">
                            <CellInput
                              value={ind.reference_range}
                              aria-label="Референсный диапазон"
                              onChange={(e) =>
                                updateIndicatorLocal(ind.id, {
                                  reference_range: e.target.value,
                                })
                              }
                            />
                          </TableCell>
                          <TableCell className="p-1">
                            <CellInput
                              type="date"
                              value={ind.test_date ?? ""}
                              aria-label="Дата анализа"
                              onChange={(e) =>
                                updateIndicatorLocal(ind.id, { test_date: e.target.value })
                              }
                            />
                          </TableCell>
                          <TableCell className="p-2">
                            <VerificationChip confirmed={confirmed} />
                          </TableCell>
                          <TableCell className="p-2">
                            <div className="flex items-center justify-end gap-1">
                              <button
                                onClick={() => saveIndicator(ind)}
                                className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-primary"
                                aria-label="Сохранить изменения"
                                title="Сохранить"
                              >
                                <Save className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() =>
                                  setStatus(ind, confirmed ? "needs_review" : "confirmed")
                                }
                                className={`cursor-pointer rounded-md p-1.5 transition-colors hover:bg-muted ${
                                  confirmed
                                    ? "text-muted-foreground hover:text-warning"
                                    : "text-muted-foreground hover:text-success"
                                }`}
                                aria-label={confirmed ? "Снять подтверждение" : "Подтвердить"}
                                title={confirmed ? "Снять подтверждение" : "Подтвердить"}
                              >
                                {confirmed ? <X className="h-4 w-4" /> : <Check className="h-4 w-4" />}
                              </button>
                              <button
                                onClick={() => removeIndicator(ind.id)}
                                className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-destructive"
                                aria-label="Удалить показатель"
                                title="Удалить"
                              >
                                <Trash2 className="h-4 w-4" />
                              </button>
                            </div>
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              )}
            </div>

            {indicators.length > 0 && (
              <div className="flex items-center gap-2 border-t border-border bg-muted/30 px-5 py-3 text-xs text-muted-foreground">
                <ShieldCheck className="h-3.5 w-3.5 text-success" />
                В сводку попадут только подтверждённые значения
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
