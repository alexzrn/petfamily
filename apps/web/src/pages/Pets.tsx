import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api } from "../api";
import type { Pet } from "../api";
import { useAuth } from "../auth";

const inputStyle = { width: "100%" };

export default function Pets() {
  const { user, logout } = useAuth();
  const [pets, setPets] = useState<Pet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);
  const [name, setName] = useState("");
  const [species, setSpecies] = useState("");
  const [sex, setSex] = useState("");
  const [birthYear, setBirthYear] = useState("");
  const [breed, setBreed] = useState("");
  const [traits, setTraits] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      setPets(await api.listPets());
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось загрузить питомцев");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function submit(e: FormEvent) {
    e.preventDefault();
    const petName = name.trim();
    if (petName === "") {
      setFormError("Кличка обязательна");
      return;
    }
    let year: number | null = null;
    if (birthYear.trim() !== "") {
      const n = Number(birthYear.trim());
      if (!Number.isInteger(n) || n < 1980 || n > 2100) {
        setFormError("Год рождения — число от 1980 до 2100");
        return;
      }
      year = n;
    }
    setFormError("");
    setBusy(true);
    try {
      const created = await api.createPet({
        name: petName,
        species: species.trim() || undefined,
        sex: sex.trim() || undefined,
        birth_year: year,
        breed: breed.trim() || undefined,
        traits: traits.trim() || undefined,
      });
      setPets((prev) => [...prev, created]);
      setName("");
      setSpecies("");
      setSex("");
      setBirthYear("");
      setBreed("");
      setTraits("");
    } catch (err) {
      setFormError(err instanceof Error ? err.message : "Не удалось создать питомца");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section>
      <h1>Питомцы</h1>
      {user ? (
        <p>
          {user.display_name} ({user.email}){" "}
          <button type="button" onClick={() => void logout()}>
            Выйти
          </button>
        </p>
      ) : null}

      <h2>Новый питомец</h2>
      <form onSubmit={submit} style={{ display: "grid", gap: "0.5rem", maxWidth: 320 }}>
        <label>
          Кличка *
          <input value={name} onChange={(e) => setName(e.target.value)} style={inputStyle} />
        </label>
        <label>
          Вид
          <input value={species} onChange={(e) => setSpecies(e.target.value)} style={inputStyle} />
        </label>
        <label>
          Пол
          <input value={sex} onChange={(e) => setSex(e.target.value)} style={inputStyle} />
        </label>
        <label>
          Год рождения
          <input
            value={birthYear}
            onChange={(e) => setBirthYear(e.target.value)}
            inputMode="numeric"
            style={inputStyle}
          />
        </label>
        <label>
          Порода
          <input value={breed} onChange={(e) => setBreed(e.target.value)} style={inputStyle} />
        </label>
        <label>
          Особенности
          <input value={traits} onChange={(e) => setTraits(e.target.value)} style={inputStyle} />
        </label>
        {formError ? (
          <p role="alert" style={{ color: "crimson" }}>
            {formError}
          </p>
        ) : null}
        <button type="submit" disabled={busy}>
          {busy ? "Сохраняем…" : "Добавить питомца"}
        </button>
      </form>

      <h2>Мои питомцы</h2>
      {loading ? (
        <p>Загружаю…</p>
      ) : error ? (
        <p role="alert" style={{ color: "crimson" }}>
          {error}{" "}
          <button type="button" onClick={() => void load()}>
            Повторить
          </button>
        </p>
      ) : pets.length === 0 ? (
        <p>Пока никого — добавьте первого выше.</p>
      ) : (
        <ul style={{ display: "grid", gap: "0.5rem", padding: 0, listStyle: "none" }}>
          {pets.map((p) => (
            <li
              key={p.id}
              style={{ border: "1px solid #ddd", borderRadius: 8, padding: "0.5rem 0.75rem" }}
            >
              <strong>{p.name}</strong>
              {[p.species, p.sex, p.breed].filter(Boolean).join(" · ") || null}
              {p.birth_year ? <span> · {p.birth_year} г.</span> : null}
              {p.traits ? <div>{p.traits}</div> : null}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
