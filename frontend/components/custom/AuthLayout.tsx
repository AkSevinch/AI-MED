import { ShieldCheck } from "lucide-react";
import ThemeToggle from "@/components/custom/ThemeToggle";

function BrandMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" className={className} aria-hidden="true">
      <rect width="32" height="32" rx="9" fill="#ffffff" />
      <path
        d="M9 16h3.2l1.9-5 3.4 10 2-6.2 1.4 1.2H23"
        fill="none"
        stroke="#0e7490"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export default function AuthLayout({
  title,
  subtitle,
  points,
  children,
}: {
  title: string;
  subtitle: string;
  points: string[];
  children: React.ReactNode;
}) {
  return (
    <div className="relative grid min-h-screen lg:grid-cols-2">
      <div className="absolute right-5 top-5 z-10">
        <ThemeToggle />
      </div>
      {/* Брендовая панель */}
      <div className="relative hidden flex-col justify-between bg-gradient-to-br from-primary to-[#075985] p-10 text-white lg:flex">
        <div className="flex items-center gap-3">
          <BrandMark className="h-9 w-9" />
          <div className="leading-tight">
            <p className="font-display text-lg font-semibold">AI MED</p>
            <p className="text-xs text-white/70">
              Единая сводка сердечно-сосудистого здоровья
            </p>
          </div>
        </div>

        <div className="max-w-md space-y-6">
          <h2 className="font-display text-3xl font-semibold leading-snug">
            {title}
          </h2>
          <ul className="space-y-3">
            {points.map((p) => (
              <li key={p} className="flex items-start gap-3 text-sm text-white/85">
                <ShieldCheck className="mt-0.5 h-4 w-4 flex-shrink-0 text-white" />
                <span className="leading-relaxed">{p}</span>
              </li>
            ))}
          </ul>
        </div>

        <p className="text-xs text-white/60">
          Сервис носит информационный характер и не заменяет консультацию врача.
        </p>
      </div>

      {/* Панель формы */}
      <div className="flex items-center justify-center bg-background px-6 py-10 sm:px-10">
        <div className="w-full max-w-md">
          <div className="mb-8 flex items-center gap-3 lg:hidden">
            <BrandMark className="h-9 w-9" />
            <div className="leading-tight">
              <p className="font-display text-lg font-semibold text-foreground">
                AI MED
              </p>
              <p className="text-xs text-muted-foreground">{subtitle}</p>
            </div>
          </div>
          {children}
        </div>
      </div>
    </div>
  );
}
