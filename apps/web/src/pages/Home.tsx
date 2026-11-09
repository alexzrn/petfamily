import { Link } from "react-router-dom";

export default function Home() {
  return (
    <section>
      <h1>Petfamily</h1>
      <p>Каркас фронта: роутинг работает, страницы — заглушки.</p>
      <p>
        <Link to="/pets">Питомцы</Link> · <Link to="/map">Карта</Link>
      </p>
    </section>
  );
}
