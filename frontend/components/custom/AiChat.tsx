"use client";

import { useEffect, useRef, useState } from "react";
import { Send, Sparkles, Bot, User, AlertTriangle, ShieldCheck } from "lucide-react";
import { api, type ChatMessage, type ChatResponse } from "@/lib/api";

const SUGGESTIONS = [
  "Что отклонилось от нормы?",
  "Какие вопросы задать врачу?",
  "Объясни мои последние показатели",
  "Что изменилось с прошлого анализа?",
];

export function AiChat({ className }: { className?: string }) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({
      top: scrollRef.current.scrollHeight,
      behavior: "smooth",
    });
  }, [messages, busy]);

  const send = async (text: string) => {
    const question = text.trim();
    if (!question || busy) return;

    const history = messages.slice(-10);
    setMessages((prev) => [...prev, { role: "user", content: question }]);
    setInput("");
    setBusy(true);

    try {
      const res = await api.post<ChatResponse>("/api/chat", {
        question,
        history,
      });
      setMessages((prev) => [...prev, { role: "assistant", content: res.answer }]);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Не удалось получить ответ";
      setMessages((prev) => [
        ...prev,
        { role: "assistant", content: `__error__${message}` },
      ]);
    } finally {
      setBusy(false);
    }
  };

  const empty = messages.length === 0;

  return (
    <div className={`surface flex flex-col overflow-hidden ${className ?? ""}`}>
      <div className="flex items-center gap-2.5 border-b border-border px-5 py-4">
        <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
          <Bot className="h-4 w-4 text-primary" />
        </span>
        <div>
          <h2 className="font-display font-semibold text-foreground">
            ИИ-ассистент
          </h2>
          <p className="text-xs text-muted-foreground">
            Отвечает по вашим подтверждённым данным
          </p>
        </div>
      </div>

      <div
        ref={scrollRef}
        className="hide-scrollbar flex-1 space-y-4 overflow-y-auto px-5 py-4"
      >
        {empty && (
          <div className="flex flex-col items-center gap-4 py-6 text-center">
            <span className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
              <Sparkles className="h-5 w-5 text-primary" />
            </span>
            <p className="max-w-sm text-sm text-muted-foreground">
              Спросите о своих показателях, отклонениях от референса или о том,
              что обсудить с врачом.
            </p>
            <div className="flex flex-wrap justify-center gap-2">
              {SUGGESTIONS.map((s) => (
                <button
                  key={s}
                  type="button"
                  onClick={() => void send(s)}
                  className="cursor-pointer rounded-full border border-border bg-background px-3 py-1.5 text-xs font-medium text-foreground transition-colors hover:border-primary/30 hover:bg-primary/5"
                >
                  {s}
                </button>
              ))}
            </div>
          </div>
        )}

        {messages.map((m, i) => {
          const isError = m.content.startsWith("__error__");
          const text = isError ? m.content.slice("__error__".length) : m.content;
          const isUser = m.role === "user";
          return (
            <div
              key={i}
              className={`flex gap-2.5 ${isUser ? "flex-row-reverse" : ""}`}
            >
              <span
                className={`flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg ${
                  isUser ? "bg-muted" : "bg-primary/10"
                }`}
              >
                {isUser ? (
                  <User className="h-3.5 w-3.5 text-muted-foreground" />
                ) : (
                  <Bot className="h-3.5 w-3.5 text-primary" />
                )}
              </span>
              <div
                className={`max-w-[80%] rounded-xl px-3.5 py-2.5 text-sm leading-relaxed ${
                  isUser
                    ? "bg-primary text-primary-foreground"
                    : isError
                      ? "border border-warning/25 bg-warning/10 text-warning"
                      : "bg-muted text-foreground"
                }`}
              >
                {isError && (
                  <span className="mb-1 flex items-center gap-1.5 text-xs font-medium">
                    <AlertTriangle className="h-3.5 w-3.5" />
                    ИИ недоступен
                  </span>
                )}
                <p className="whitespace-pre-line">{text}</p>
              </div>
            </div>
          );
        })}

        {busy && (
          <div className="flex gap-2.5">
            <span className="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10">
              <Bot className="h-3.5 w-3.5 text-primary" />
            </span>
            <div className="flex items-center gap-1 rounded-xl bg-muted px-3.5 py-3">
              <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted-foreground [animation-delay:-0.3s]" />
              <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted-foreground [animation-delay:-0.15s]" />
              <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-muted-foreground" />
            </div>
          </div>
        )}
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          void send(input);
        }}
        className="border-t border-border p-3"
      >
        <div className="flex items-end gap-2">
          <textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                void send(input);
              }
            }}
            rows={1}
            placeholder="Задайте вопрос о своём здоровье…"
            className="max-h-32 min-h-[2.5rem] flex-1 resize-none rounded-lg border border-input bg-background px-3 py-2.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-primary/40 focus:ring-2 focus:ring-ring/30"
          />
          <button
            type="submit"
            disabled={busy || !input.trim()}
            className="flex h-10 w-10 flex-shrink-0 cursor-pointer items-center justify-center rounded-lg bg-primary text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-default disabled:opacity-50"
            aria-label="Отправить"
          >
            <Send className="h-4 w-4" />
          </button>
        </div>
        <p className="mt-2 flex items-center gap-1.5 text-[0.7rem] text-muted-foreground">
          <ShieldCheck className="h-3 w-3 text-success" />
          Ассистент не ставит диагноз и не назначает лечение. При тревожных
          симптомах звоните 103 или 112.
        </p>
      </form>
    </div>
  );
}
