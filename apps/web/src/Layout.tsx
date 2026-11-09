import { Link, Outlet } from "react-router-dom";

const links = [
  { to: "/pets", label: "Питомцы" },
  { to: "/map", label: "Карта" },
  { to: "/walks", label: "Прогулки" },
  { to: "/meetups", label: "Встречи" },
  { to: "/login", label: "Вход" },
  { to: "/register", label: "Регистрация" },
];

export default function Layout() {
  return (
    <div>
      <header
        style={{
          display: "flex",
          gap: "1rem",
          alignItems: "center",
          padding: "0.75rem 1rem",
          borderBottom: "1px solid #ddd",
        }}
      >
        <Link to="/" style={{ fontWeight: 700, textDecoration: "none" }}>
          Petfamily
        </Link>
        <nav style={{ display: "flex", gap: "0.75rem" }}>
          {links.map((l) => (
            <Link key={l.to} to={l.to}>
              {l.label}
            </Link>
          ))}
        </nav>
      </header>
      <main style={{ padding: "1rem" }}>
        <Outlet />
      </main>
    </div>
  );
}
