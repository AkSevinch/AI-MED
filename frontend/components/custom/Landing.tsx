"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import "./landing.css";

const HERO_VIDEO =
  "https://jovxgdctpqexoiwfspue.supabase.co/storage/v1/object/sign/asset-media/31bc78ca-e0bf-4939-9f19-2e5badccb169/5a8c3d7d2f3a-hf_20260818_072341_50851634-bbc3-4c33-9acc-7647d4db44aa.mp4?token=eyJraWQiOiJzdG9yYWdlLXVybC1zaWduaW5nLWtleV9iMjk0Y2Y1ZS01ODRhLTQ5ODItOGY3MS04NGQzZDAwMzJiNzgiLCJhbGciOiJIUzI1NiJ9.eyJ1cmwiOiJhc3NldC1tZWRpYS8zMWJjNzhjYS1lMGJmLTQ5MzktOWYxOS0yZTViYWRjY2IxNjkvNWE4YzNkN2QyZjNhLWhmXzIwMjYwODE4XzA3MjM0MV81MDg1MTYzNC1iYmMzLTRjMzMtOWFjYy03NjQ3ZDRkYjQ0YWEubXA0Iiwic2NvcGUiOiJkb3dubG9hZCIsImlhdCI6MTc4NzQyMTk2NSwiZXhwIjoyMTAyNzgxOTY1fQ.Tgfb94VOkz2xmMRnM14C8bHEtUk3yQHsG0-0jp73cT4";

const NAV = [
  { label: "Benefits", href: "#benefits", mod: "appear--scale", d: "0.16s" },
  { label: "How It Works", href: "#how-it-works", mod: "appear--soft", d: "0.28s" },
  { label: "FAQs", href: "#faqs", mod: "appear--scale", d: "0.40s" },
  { label: "Pricing", href: "#pricing", mod: "appear--soft", d: "0.52s" },
];

export default function Landing() {
  const rootRef = useRef<HTMLDivElement>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;

    const targets = Array.from(
      root.querySelectorAll<HTMLElement>(".appear, .hero-photo"),
    );
    const onEnd = (e: AnimationEvent) => {
      (e.currentTarget as HTMLElement).classList.add("is-in");
    };
    targets.forEach((el) =>
      el.addEventListener("animationend", onEnd, { once: true }),
    );

    const raf = requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        targets.forEach((el) => {
          const anims = el.getAnimations ? el.getAnimations() : [];
          const active = anims.some(
            (a) => a.playState === "running" || a.playState === "finished",
          );
          if (!active) el.classList.add("is-in");
        });
      }),
    );

    return () => {
      targets.forEach((el) => el.removeEventListener("animationend", onEnd));
      cancelAnimationFrame(raf);
    };
  }, []);

  useEffect(() => {
    document.documentElement.lang = "en";
    return () => {
      document.documentElement.lang = "ru";
    };
  }, []);

  useEffect(() => {
    document.body.classList.toggle("menu-open", menuOpen);
    return () => document.body.classList.remove("menu-open");
  }, [menuOpen]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenuOpen(false);
    };
    const mq = window.matchMedia("(min-width: 901px)");
    const onResize = () => {
      if (mq.matches) setMenuOpen(false);
    };
    window.addEventListener("keydown", onKey);
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("keydown", onKey);
      window.removeEventListener("resize", onResize);
    };
  }, []);

  return (
    <div ref={rootRef} className={"vsp" + (menuOpen ? " menu-open" : "")}>
      <div className="grain" />

      <div className="hero-photo">
        <video
          src={HERO_VIDEO}
          autoPlay
          muted
          loop
          playsInline
          preload="auto"
        />
      </div>

      <div className="page">
        <div
          className="menu-backdrop"
          onClick={() => setMenuOpen(false)}
          aria-hidden="true"
        />

        <header className="header">
          <Link href="#top" className="logo appear appear--scale" aria-label="Vesper.ai" style={{ ["--d" as string]: "0.08s" }}>
            <svg className="logo-mark" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <g transform="rotate(-30 12 12)">
                <circle cx="7.3" cy="3.2" r="1.45" />
                <rect x="5.5" y="4.7" width="3.6" height="14.6" rx="1.8" />
                <rect x="14.9" y="4.7" width="3.6" height="14.6" rx="1.8" />
                <circle cx="16.7" cy="20.8" r="1.45" />
              </g>
            </svg>
            <span>
              Vesper<span className="logo-suffix">.ai</span>
            </span>
          </Link>

          <nav id="site-nav" aria-label="Primary">
            {NAV.map((item) => (
              <a
                key={item.label}
                href={item.href}
                className={`nav-link appear ${item.mod}`}
                style={{ ["--d" as string]: item.d }}
                onClick={() => setMenuOpen(false)}
              >
                {item.label}
              </a>
            ))}
          </nav>

          <div className="header-right">
            <Link
              href="/register"
              className="btn btn-solid header-cta appear appear--scale"
              style={{ ["--d" as string]: "0.34s" }}
            >
              Start for Free
            </Link>
            <button
              type="button"
              className="burger appear appear--scale"
              style={{ ["--d" as string]: "0.34s" }}
              aria-controls="site-nav"
              aria-expanded={menuOpen}
              aria-label={menuOpen ? "Close menu" : "Open menu"}
              onClick={() => setMenuOpen((v) => !v)}
            >
              <span />
              <span />
              <span />
            </button>
          </div>
        </header>

        <main className="hero" id="top">
          <div className="hero-copy">
            <div
              className="badge appear appear--pop"
              style={{ ["--d" as string]: "0.22s" }}
            >
              <svg className="badge-star" viewBox="0 0 24 24" fill="#fff" aria-hidden="true">
                <path d="M12 2.6C12.55 2.6 12.88 3.15 13.08 4.7c.62 4.7 1.52 5.6 6.22 6.22 1.55.2 2.1.53 2.1 1.08s-.55.88-2.1 1.08c-4.7.62-5.6 1.52-6.22 6.22-.2 1.55-.53 2.1-1.08 2.1s-.88-.55-1.08-2.1c-.62-4.7-1.52-5.6-6.22-6.22C3.15 12.88 2.6 12.55 2.6 12s.55-.88 2.1-1.08c4.7-.62 5.6-1.52 6.22-6.22C11.12 3.15 11.45 2.6 12 2.6Z" />
              </svg>
              Operational AI Infrastructure
            </div>

            <h1>
              <span
                className="headline-line appear appear--mask"
                style={{ ["--d" as string]: "0.42s" }}
              >
                Train <em>AI agents</em> on your
              </span>
              <span
                className="headline-line appear appear--mask"
                style={{ ["--d" as string]: "0.62s" }}
              >
                workflows in minutes.
              </span>
            </h1>

            <p
              className="lede appear appear--soft"
              style={{ ["--d" as string]: "0.82s" }}
            >
              Deploy adaptive AI agents that learn, execute, and scale
              operational tasks across your business.
            </p>

            <div className="hero-actions">
              <Link
                href="/register"
                className="btn btn-solid appear appear--btn"
                style={{ ["--d" as string]: "0.96s" }}
              >
                Start for Free
              </Link>
              <Link
                href="/login"
                className="btn btn-ghost appear appear--side"
                style={{ ["--d" as string]: "1.10s" }}
              >
                See it in action
              </Link>
            </div>
          </div>
        </main>

        <footer className="stats">
          <div
            className="stat appear appear--stat"
            style={{ ["--d" as string]: "1.12s" }}
          >
            <svg className="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
              <defs>
                <linearGradient id="vspPillL" x1="3" y1="2" x2="14" y2="22" gradientUnits="userSpaceOnUse">
                  <stop offset="0" stopColor="#ffffff" stopOpacity="0.38" />
                  <stop offset="1" stopColor="#3a3a3a" stopOpacity="0.62" />
                </linearGradient>
                <linearGradient id="vspPillR" x1="13" y1="2" x2="24" y2="22" gradientUnits="userSpaceOnUse">
                  <stop offset="0" stopColor="#3a3a3a" stopOpacity="0.38" />
                  <stop offset="1" stopColor="#ffffff" stopOpacity="0.62" />
                </linearGradient>
              </defs>
              <rect x="3.4" y="2.6" width="7.2" height="18.8" rx="3.6" fill="url(#vspPillL)" />
              <rect x="13.4" y="2.6" width="7.2" height="18.8" rx="3.6" fill="url(#vspPillR)" />
              <rect x="9.2" y="10.9" width="5.6" height="2.2" rx="1.1" fill="#4a4a4a" />
            </svg>
            4.2M+ workflows automated
          </div>

          <div
            className="stat appear appear--stat"
            style={{ ["--d" as string]: "1.28s" }}
          >
            <svg className="stat-icon" viewBox="0 0 24 24" aria-hidden="true">
              <rect x="2.4" y="2.4" width="19.2" height="19.2" rx="6.2" fill="#ffffff" />
              <g fill="none" stroke="#111" strokeWidth="1.85" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 7.1v7.4" />
                <path d="M8.15 12.35L12 16.2l3.85-3.85" />
              </g>
            </svg>
            92% reduction in manual operations
          </div>

          <div
            className="stat appear appear--stat"
            style={{ ["--d" as string]: "1.44s" }}
          >
            <svg className="stat-icon-wide" viewBox="0 0 40 22" aria-hidden="true">
              <g>
                <circle cx="10.2" cy="11" r="9.2" fill="#2b2b2b" />
                <path d="M6.4 4.6 L7.7 1.5 L9.4 4.3 Z" fill="#2b2b2b" />
                <path d="M11 4.6 L12.7 1.5 L14.4 4.3 Z" fill="#2b2b2b" />
                <ellipse cx="10.2" cy="12.1" rx="4.15" ry="3.7" fill="#f4f4f4" />
                <circle cx="8.7" cy="11.6" r="0.7" fill="#1a1a1a" />
                <circle cx="11.7" cy="11.6" r="0.7" fill="#1a1a1a" />
              </g>
              <g>
                <circle cx="20.2" cy="11" r="9.2" fill="#ffffff" />
                <circle cx="18" cy="10" r="1.7" fill="#111" />
                <circle cx="22.4" cy="10" r="1.7" fill="#111" />
                <ellipse cx="20.2" cy="12" rx="0.95" ry="1.35" fill="#111" />
                <path d="M17.6 13.6 Q20.2 15.8 22.8 13.6" fill="none" stroke="#111" strokeWidth="1.2" strokeLinecap="round" />
              </g>
              <g>
                <circle cx="30.2" cy="11" r="9.2" fill="#f26b1d" />
                <text x="30.2" y="15.1" textAnchor="middle" fontSize="12.5" fontWeight="700" fill="#ffffff" fontFamily="Inter, system-ui, sans-serif">
                  e
                </text>
              </g>
            </svg>
            180+ operational teams onboarded
          </div>
        </footer>
      </div>
    </div>
  );
}
