import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { getToken } from "./api";

// Без токена — на /login. Проверка лёгкая, без запроса к API.
export default function RequireAuth({ children }: { children: ReactNode }) {
  if (!getToken()) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
