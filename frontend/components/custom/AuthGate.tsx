"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { usePathname, useRouter } from "next/navigation";
import { api, type MeResponse } from "@/lib/api";

type AuthState = {
  me: MeResponse | null;
  loading: boolean;
  refresh: () => Promise<MeResponse | null>;
};

const AuthContext = createContext<AuthState>({
  me: null,
  loading: true,
  refresh: async () => null,
});

export const useAuth = () => useContext(AuthContext);

const PUBLIC_PATHS = ["/login", "/register", "/"];
const AUTHED_REDIRECT = ["/login", "/register", "/consent"];

export default function AuthGate({ children }: { children: React.ReactNode }) {
  const [me, setMe] = useState<MeResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const pathname = usePathname();
  const router = useRouter();

  const refresh = useCallback(async () => {
    try {
      const m = await api.get<MeResponse>("/api/auth/me");
      setMe(m);
      return m;
    } catch {
      const anon: MeResponse = { authenticated: false };
      setMe(anon);
      return anon;
    }
  }, []);

  useEffect(() => {
    refresh().finally(() => setLoading(false));
  }, [refresh]);

  const authed = !!me?.authenticated;
  const consented = !!me?.has_consent;

  let allowed = false;
  let redirectTo: string | null = null;

  if (!authed) {
    allowed = PUBLIC_PATHS.includes(pathname);
    redirectTo = allowed ? null : "/login";
  } else if (!consented) {
    allowed = pathname === "/consent";
    redirectTo = allowed ? null : "/consent";
  } else {
    allowed = !AUTHED_REDIRECT.includes(pathname);
    redirectTo = allowed ? null : "/";
  }

  useEffect(() => {
    if (!loading && redirectTo) router.replace(redirectTo);
  }, [loading, redirectTo, router]);

  return (
    <AuthContext.Provider value={{ me, loading, refresh }}>
      {loading || !allowed ? (
        <div className="flex min-h-screen items-center justify-center bg-background">
          <div className="h-10 w-10 animate-spin rounded-full border-4 border-primary/20 border-t-primary" />
        </div>
      ) : (
        children
      )}
    </AuthContext.Provider>
  );
}
