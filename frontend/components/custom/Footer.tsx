import { HeartPulse, ShieldCheck, Info } from "lucide-react";

const Footer: React.FC = () => {
  return (
    <footer className="mt-16 border-t border-border bg-card">
      <div className="mx-auto grid max-w-6xl grid-cols-1 gap-8 px-6 py-10 md:grid-cols-3">
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
              <HeartPulse className="h-4 w-4 text-primary" />
            </span>
            <span className="font-display font-semibold text-foreground">
              AI MED
            </span>
          </div>
          <p className="max-w-xs text-sm leading-relaxed text-muted-foreground">
            Единая сводка сердечно-сосудистого здоровья: анализы, измерения и
            история — в одном месте.
          </p>
        </div>

        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Info className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-semibold text-foreground">Важно</span>
          </div>
          <p className="max-w-xs text-sm leading-relaxed text-muted-foreground">
            Сервис не ставит диагнозы, не назначает лечение и не заменяет
            консультацию врача. Сводка носит информационный характер.
          </p>
        </div>

        <div className="space-y-3 md:text-right">
          <div className="flex items-center gap-2 md:justify-end">
            <ShieldCheck className="h-4 w-4 text-primary" />
            <span className="text-sm font-semibold text-foreground">
              Приватность
            </span>
          </div>
          <p className="text-sm leading-relaxed text-muted-foreground md:ml-auto md:max-w-xs">
            Данные привязаны к вашему аккаунту и видны только вам.
          </p>
          <p className="pt-2 text-xs text-muted-foreground">
            © {new Date().getFullYear()} AI MED
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
