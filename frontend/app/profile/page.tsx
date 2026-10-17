"use client";

import React, { useCallback, useEffect, useState } from "react";
import {
  UserRound,
  Activity,
  Heart,
  FileText,
  Plus,
  Trash2,
  Calendar,
  Pill,
  Stethoscope,
  ClipboardList,
} from "lucide-react";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type Questionnaire = {
  birth_date: string;
  sex: string;
  known_diagnoses: string;
  current_medications: string;
  cardiovascular_history: string;
  weight_kg: number | null;
  risk_factors: string;
  updated_at?: string;
};

type Diagnosis = {
  id: number;
  name: string;
  comment: string;
};

type Medication = {
  id: number;
  name: string;
  drug_class: string;
  frequency: string;
  status: string;
  comment: string;
};

type Measurement = {
  id: number;
  kind: string;
  measured_at: string;
  systolic: number | null;
  diastolic: number | null;
  pulse: number | null;
  weight: number | null;
  comment: string;
};

const EMPTY_QUESTIONNAIRE: Questionnaire = {
  birth_date: "",
  sex: "",
  known_diagnoses: "",
  current_medications: "",
  cardiovascular_history: "",
  weight_kg: null,
  risk_factors: "",
};

const TABS = [
  { id: "questionnaire", label: "Анкета", icon: ClipboardList },
  { id: "diagnoses", label: "Диагнозы", icon: Stethoscope },
  { id: "medications", label: "Лекарства", icon: Pill },
  { id: "measurements", label: "Измерения", icon: Activity },
] as const;

type TabId = (typeof TABS)[number]["id"];

const MED_STATUS: Record<string, string> = {
  taking: "Принимаю",
  paused: "Приостановлено",
  unknown: "Неизвестно",
};

const KIND_LABEL: Record<string, string> = {
  pressure: "Давление",
  pulse: "Пульс",
  weight: "Вес",
};

const FIELD_CLASS =
  "h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50";

export default function ProfilePage() {
  const [tab, setTab] = useState<TabId>("questionnaire");

  return (
    <div className="mx-auto max-w-5xl px-6 py-8">
      <div className="mb-8">
        <p className="mb-1 text-sm font-medium text-primary">Личный кабинет</p>
        <h1 className="font-display text-3xl font-semibold text-foreground">
          Профиль здоровья
        </h1>
        <p className="mt-2 max-w-2xl text-muted-foreground">
          Данные, из которых формируется ваша единая сводка здоровья.
        </p>
      </div>

      <div className="mb-6 inline-flex flex-wrap gap-1 rounded-xl border border-border bg-card p-1">
        {TABS.map((t) => {
          const Icon = t.icon;
          const active = tab === t.id;
          return (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={`flex cursor-pointer items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors duration-200 ${
                active
                  ? "bg-primary text-primary-foreground shadow-sm"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
              }`}
            >
              <Icon className="h-4 w-4" />
              {t.label}
            </button>
          );
        })}
      </div>

      {tab === "questionnaire" && <QuestionnaireSection />}
      {tab === "diagnoses" && <DiagnosesSection />}
      {tab === "medications" && <MedicationsSection />}
      {tab === "measurements" && <MeasurementsSection />}
    </div>
  );
}

function SectionCard({
  title,
  subtitle,
  icon: Icon,
  children,
}: {
  title: string;
  subtitle: string;
  icon: React.ComponentType<{ className?: string }>;
  children: React.ReactNode;
}) {
  return (
    <div className="surface overflow-hidden">
      <div className="flex items-center gap-3 border-b border-border px-6 py-4">
        <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Icon className="h-[18px] w-[18px]" />
        </span>
        <div>
          <h2 className="font-display font-semibold text-foreground">{title}</h2>
          <p className="text-xs text-muted-foreground">{subtitle}</p>
        </div>
      </div>
      <div className="p-6">{children}</div>
    </div>
  );
}

function ErrorNote({ message }: { message: string }) {
  if (!message) return null;
  return (
    <p className="rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2 text-sm text-destructive">
      {message}
    </p>
  );
}

function EmptyNote({ text }: { text: string }) {
  return (
    <div className="rounded-xl border border-dashed border-border py-8 text-center text-sm text-muted-foreground">
      {text}
    </div>
  );
}

function AddPanel({ children }: { children: React.ReactNode }) {
  return (
    <div className="mb-6 space-y-4 rounded-xl border border-border bg-muted/40 p-4">
      {children}
    </div>
  );
}

// --- FR2: анкета ---

function QuestionnaireSection() {
  const [form, setForm] = useState<Questionnaire>(EMPTY_QUESTIONNAIRE);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api
      .get<Questionnaire>("/api/profile/questionnaire")
      .then((data) => setForm({ ...EMPTY_QUESTIONNAIRE, ...data }))
      .catch((e) => setError((e as Error).message))
      .finally(() => setLoading(false));
  }, []);

  const set = (patch: Partial<Questionnaire>) =>
    setForm((prev) => ({ ...prev, ...patch }));

  const save = async () => {
    setError("");
    setSaving(true);
    setSaved(false);
    try {
      const data = await api.put<Questionnaire>("/api/profile/questionnaire", {
        ...form,
        birth_date: form.birth_date,
        weight_kg: form.weight_kg,
      });
      setForm({ ...EMPTY_QUESTIONNAIRE, ...data });
      setSaved(true);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <SectionCard
      title="Анкета контекста"
      subtitle="Общая информация о вашем здоровье"
      icon={ClipboardList}
    >
      {loading ? (
        <p className="text-sm text-muted-foreground">Загрузка…</p>
      ) : (
        <>
          <div className="grid gap-5 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="birth_date">Дата рождения</Label>
              <Input
                id="birth_date"
                type="date"
                className={FIELD_CLASS}
                value={form.birth_date}
                onChange={(e) => set({ birth_date: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="sex">Пол</Label>
              <select
                id="sex"
                value={form.sex}
                onChange={(e) => set({ sex: e.target.value })}
                className={FIELD_CLASS}
              >
                <option value="">Не указан</option>
                <option value="male">Мужской</option>
                <option value="female">Женский</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="weight_kg">Вес, кг</Label>
              <Input
                id="weight_kg"
                type="number"
                step="0.1"
                className={FIELD_CLASS}
                value={form.weight_kg ?? ""}
                onChange={(e) =>
                  set({ weight_kg: e.target.value === "" ? null : Number(e.target.value) })
                }
              />
            </div>
          </div>

          <div className="mt-6 space-y-5">
            {(
              [
                ["known_diagnoses", "Известные диагнозы"],
                ["current_medications", "Текущие лекарства"],
                ["cardiovascular_history", "Сердечно-сосудистый анамнез"],
                ["risk_factors", "Факторы риска"],
              ] as const
            ).map(([key, label]) => (
              <div key={key} className="space-y-2">
                <Label htmlFor={key}>{label}</Label>
                <textarea
                  id={key}
                  rows={2}
                  value={form[key]}
                  onChange={(e) => set({ [key]: e.target.value } as Partial<Questionnaire>)}
                  className="w-full resize-none rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
                />
              </div>
            ))}
          </div>

          <div className="mt-6 flex items-center gap-4">
            <Button onClick={save} disabled={saving}>
              {saving ? "Сохранение…" : "Сохранить"}
            </Button>
            {saved && (
              <span className="flex items-center gap-1 text-sm text-success">
                <Heart className="h-4 w-4" /> Сохранено
              </span>
            )}
          </div>
          <div className="mt-4">
            <ErrorNote message={error} />
          </div>
        </>
      )}
    </SectionCard>
  );
}

// --- FR3: диагнозы ---

function DiagnosesSection() {
  const [items, setItems] = useState<Diagnosis[]>([]);
  const [name, setName] = useState("");
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await api.get<Diagnosis[]>("/api/diagnoses"));
    } catch (e) {
      setError((e as Error).message);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const add = async () => {
    setError("");
    if (!name.trim()) return;
    try {
      await api.post("/api/diagnoses", { name, comment });
      setName("");
      setComment("");
      await load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const remove = async (id: number) => {
    try {
      await api.del(`/api/diagnoses/${id}`);
      await load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <SectionCard
      title="Диагнозы"
      subtitle="Диагнозы, установленные вашим врачом"
      icon={Stethoscope}
    >
      <AddPanel>
        <div className="grid gap-3 md:grid-cols-2">
          <Input
            className={FIELD_CLASS}
            placeholder="Название диагноза"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <Input
            className={FIELD_CLASS}
            placeholder="Комментарий (необязательно)"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
          />
        </div>
        <Button onClick={add} className="w-full md:w-auto">
          <Plus className="h-4 w-4" /> Добавить диагноз
        </Button>
      </AddPanel>

      <ErrorNote message={error} />

      <div className="mt-4 space-y-2">
        {items.length === 0 && <EmptyNote text="Пока нет записей" />}
        {items.map((d) => (
          <div
            key={d.id}
            className="flex items-start justify-between gap-4 rounded-xl border border-border bg-card p-4"
          >
            <div className="flex items-start gap-3">
              <span className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                <Stethoscope className="h-4 w-4 text-primary" />
              </span>
              <div>
                <h3 className="font-medium text-foreground">{d.name}</h3>
                {d.comment && (
                  <p className="mt-0.5 text-sm text-muted-foreground">{d.comment}</p>
                )}
              </div>
            </div>
            <button
              onClick={() => remove(d.id)}
              className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-destructive"
              aria-label="Удалить"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </SectionCard>
  );
}

// --- FR3: лекарства ---

function MedicationsSection() {
  const [items, setItems] = useState<Medication[]>([]);
  const [name, setName] = useState("");
  const [drugClass, setDrugClass] = useState("");
  const [frequency, setFrequency] = useState("");
  const [status, setStatus] = useState("taking");
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await api.get<Medication[]>("/api/medications"));
    } catch (e) {
      setError((e as Error).message);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const add = async () => {
    setError("");
    if (!name.trim()) return;
    try {
      await api.post("/api/medications", {
        name,
        drug_class: drugClass,
        frequency,
        status,
        comment,
      });
      setName("");
      setDrugClass("");
      setFrequency("");
      setStatus("taking");
      setComment("");
      await load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const remove = async (id: number) => {
    try {
      await api.del(`/api/medications/${id}`);
      await load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <SectionCard
      title="Лекарства"
      subtitle="Что вы принимаете — для контекста сводки"
      icon={Pill}
    >
      <AddPanel>
        <div className="grid gap-3 md:grid-cols-2">
          <Input
            className={FIELD_CLASS}
            placeholder="Название"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <Input
            className={FIELD_CLASS}
            placeholder="Класс препарата"
            value={drugClass}
            onChange={(e) => setDrugClass(e.target.value)}
          />
          <Input
            className={FIELD_CLASS}
            placeholder="Частота приёма"
            value={frequency}
            onChange={(e) => setFrequency(e.target.value)}
          />
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={FIELD_CLASS}
          >
            <option value="taking">Принимаю</option>
            <option value="paused">Приостановлено</option>
            <option value="unknown">Неизвестно</option>
          </select>
        </div>
        <Button onClick={add} className="w-full md:w-auto">
          <Plus className="h-4 w-4" /> Добавить лекарство
        </Button>
      </AddPanel>

      <ErrorNote message={error} />

      <div className="mt-4 space-y-2">
        {items.length === 0 && <EmptyNote text="Пока нет записей" />}
        {items.map((m) => (
          <div
            key={m.id}
            className="flex items-start justify-between gap-4 rounded-xl border border-border bg-card p-4"
          >
            <div className="flex items-start gap-3">
              <span className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                <Pill className="h-4 w-4 text-primary" />
              </span>
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="font-medium text-foreground">{m.name}</h3>
                  <span className="status-chip border-primary/20 bg-primary/10 text-primary">
                    {MED_STATUS[m.status] ?? m.status}
                  </span>
                </div>
                <p className="mt-0.5 text-sm text-muted-foreground">
                  {[m.drug_class, m.frequency].filter(Boolean).join(" · ")}
                </p>
                {m.comment && (
                  <p className="mt-0.5 text-sm text-muted-foreground">{m.comment}</p>
                )}
              </div>
            </div>
            <button
              onClick={() => remove(m.id)}
              className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-destructive"
              aria-label="Удалить"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </SectionCard>
  );
}

// --- FR4: измерения ---

function MeasurementsSection() {
  const [items, setItems] = useState<Measurement[]>([]);
  const [kind, setKind] = useState("pressure");
  const [measuredAt, setMeasuredAt] = useState("");
  const [systolic, setSystolic] = useState("");
  const [diastolic, setDiastolic] = useState("");
  const [pulse, setPulse] = useState("");
  const [weight, setWeight] = useState("");
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setItems(await api.get<Measurement[]>("/api/measurements"));
    } catch (e) {
      setError((e as Error).message);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const add = async () => {
    setError("");
    const body: Record<string, unknown> = { kind, measured_at: measuredAt, comment };
    if (kind === "pressure") {
      body.systolic = Number(systolic);
      body.diastolic = Number(diastolic);
      if (pulse) body.pulse = Number(pulse);
    } else if (kind === "pulse") {
      body.pulse = Number(pulse);
    } else if (kind === "weight") {
      body.weight = Number(weight);
    }
    try {
      await api.post("/api/measurements", body);
      setMeasuredAt("");
      setSystolic("");
      setDiastolic("");
      setPulse("");
      setWeight("");
      setComment("");
      await load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const remove = async (id: number) => {
    try {
      await api.del(`/api/measurements/${id}`);
      await load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const format = (m: Measurement) => {
    if (m.kind === "pressure") {
      return `${m.systolic ?? "-"}/${m.diastolic ?? "-"} мм рт.ст.${m.pulse ? `, пульс ${m.pulse}` : ""}`;
    }
    if (m.kind === "pulse") return `${m.pulse ?? "-"} уд/мин`;
    if (m.kind === "weight") return `${m.weight ?? "-"} кг`;
    return "";
  };

  return (
    <SectionCard title="Измерения" subtitle="Давление, пульс и вес" icon={Activity}>
      <AddPanel>
        <div className="grid gap-3 md:grid-cols-2">
          <select
            value={kind}
            onChange={(e) => setKind(e.target.value)}
            className={FIELD_CLASS}
          >
            <option value="pressure">Давление</option>
            <option value="pulse">Пульс</option>
            <option value="weight">Вес</option>
          </select>
          <Input
            type="date"
            className={FIELD_CLASS}
            value={measuredAt}
            onChange={(e) => setMeasuredAt(e.target.value)}
          />
          {kind === "pressure" && (
            <>
              <Input
                type="number"
                className={FIELD_CLASS}
                placeholder="Систолическое"
                value={systolic}
                onChange={(e) => setSystolic(e.target.value)}
              />
              <Input
                type="number"
                className={FIELD_CLASS}
                placeholder="Диастолическое"
                value={diastolic}
                onChange={(e) => setDiastolic(e.target.value)}
              />
              <Input
                type="number"
                className={FIELD_CLASS}
                placeholder="Пульс (необязательно)"
                value={pulse}
                onChange={(e) => setPulse(e.target.value)}
              />
            </>
          )}
          {kind === "pulse" && (
            <Input
              type="number"
              className={FIELD_CLASS}
              placeholder="Пульс"
              value={pulse}
              onChange={(e) => setPulse(e.target.value)}
            />
          )}
          {kind === "weight" && (
            <Input
              type="number"
              step="0.1"
              className={FIELD_CLASS}
              placeholder="Вес, кг"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
            />
          )}
        </div>
        <Button onClick={add} className="w-full md:w-auto">
          <Plus className="h-4 w-4" /> Добавить измерение
        </Button>
      </AddPanel>

      <ErrorNote message={error} />

      <div className="mt-4 space-y-2">
        {items.length === 0 && <EmptyNote text="Пока нет записей" />}
        {items.map((m) => (
          <div
            key={m.id}
            className="flex items-start justify-between gap-4 rounded-xl border border-border bg-card p-4"
          >
            <div className="flex items-start gap-3">
              <span className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-lg bg-success/10">
                <FileText className="h-4 w-4 text-success" />
              </span>
              <div>
                <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  {KIND_LABEL[m.kind] ?? m.kind}
                </h3>
                <p className="mt-0.5 font-display text-lg font-semibold text-foreground">
                  {format(m)}
                </p>
                <div className="mt-0.5 flex items-center gap-1.5">
                  <Calendar className="h-3 w-3 text-muted-foreground" />
                  <span className="text-xs text-muted-foreground">{m.measured_at}</span>
                </div>
                {m.comment && (
                  <p className="mt-0.5 text-sm text-muted-foreground">{m.comment}</p>
                )}
              </div>
            </div>
            <button
              onClick={() => remove(m.id)}
              className="cursor-pointer rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-destructive"
              aria-label="Удалить"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </SectionCard>
  );
}
