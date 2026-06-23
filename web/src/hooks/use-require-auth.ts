"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuth } from "@/contexts/auth-context";
import type { Role } from "@/lib/types";

// Centralizes the per-page auth guard: redirects unauthenticated users to
// /login and wrong-role users to the dashboard. `ready` is true only once the
// caller is allowed to see the page.
export function useRequireAuth(role?: Role) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (loading) return;
    if (!user) router.replace("/login");
    else if (role && user.role !== role) router.replace("/");
  }, [loading, user, role, router]);

  const ready = !loading && !!user && (!role || user.role === role);
  return { user, ready };
}
