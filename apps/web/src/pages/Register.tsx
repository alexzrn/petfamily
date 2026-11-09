import { useState } from "react";
import type { FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../auth";

function validEmail(v: string): boolean {
  return v.length >= 3 && v.includes("@");
}

export default function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(e: FormEvent) {
    e.preventDefault();
    const mail = email.trim();
    const name = displayName.trim();
    if (!validEmail(mail)) {
      setError("Введите корректный email");
      return;
    }
    if (name === "") {
      setError("Представьтесь, пожалуйста");
      return;
    }
    if (password.length < 8) {
      setError("Пароль должен быть не короче 8 символов");
      return;
    }
    setError("");
    setBusy(true);
    try {
      await register(mail, password, name);
      navigate("/pets", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось зарегистрироваться");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section>
      <h1>Регистрация</h1>
      <form onSubmit={submit} style={{ display: "grid", gap: "0.5rem", maxWidth: 320 }}>
        <label>
          Email
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
            style={{ width: "100%" }}
          />
        </label>
        <label>
          Как вас зовут
          <input
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            autoComplete="nickname"
            style={{ width: "100%" }}
          />
        </label>
        <label>
          Пароль (минимум 8 символов)
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="new-password"
            style={{ width: "100%" }}
          />
        </label>
        {error ? (
          <p role="alert" style={{ color: "crimson" }}>
            {error}
          </p>
        ) : null}
        <button type="submit" disabled={busy}>
          {busy ? "Регистрируем…" : "Зарегистрироваться"}
        </button>
      </form>
      <p>
        Уже есть аккаунт? <Link to="/login">Войти</Link>
      </p>
    </section>
  );
}
