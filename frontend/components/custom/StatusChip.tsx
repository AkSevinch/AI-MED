import type { ComponentType } from "react";
import { cn } from "@/lib/utils";
import { CheckCircle2, Clock, AlertTriangle, Loader2 } from "lucide-react";

type DocStatus = "processing" | "needs_review" | "saved" | "failed";

const DOC: Record<
  string,
  { label: string; cls: string; icon: ComponentType<{ className?: string }> }
> = {
  processing: {
    label: "Обработка",
    cls: "border-primary/20 bg-primary/10 text-primary",
    icon: Loader2,
  },
  needs_review: {
    label: "Требует проверки",
    cls: "border-warning/25 bg-warning/10 text-warning",
    icon: Clock,
  },
  saved: {
    label: "Сохранено",
    cls: "border-success/25 bg-success/10 text-success",
    icon: CheckCircle2,
  },
  failed: {
    label: "Ошибка",
    cls: "border-destructive/25 bg-destructive/10 text-destructive",
    icon: AlertTriangle,
  },
};

export function DocStatusChip({
  status,
  className,
}: {
  status: string;
  className?: string;
}) {
  const s = DOC[status] ?? {
    label: status,
    cls: "border-border bg-muted text-muted-foreground",
    icon: Clock,
  };
  const Icon = s.icon;
  return (
    <span className={cn("status-chip", s.cls, className)}>
      <Icon className={cn("h-3 w-3", status === "processing" && "animate-spin")} />
      {s.label}
    </span>
  );
}

export function VerificationChip({
  confirmed,
  className,
}: {
  confirmed: boolean;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "status-chip",
        confirmed
          ? "border-success/25 bg-success/10 text-success"
          : "border-warning/25 bg-warning/10 text-warning",
        className,
      )}
    >
      {confirmed ? (
        <CheckCircle2 className="h-3 w-3" />
      ) : (
        <Clock className="h-3 w-3" />
      )}
      {confirmed ? "Подтверждено" : "Требует проверки"}
    </span>
  );
}

export type { DocStatus };
