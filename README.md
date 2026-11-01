# petfamily

Фронтенд проекта «Моя стая» (проект FamilyOS 12) — семейный центр заботы о собаке.

- Стек: Astro 7, Node >= 22.23.4
- Сборка: `npm ci && npm run build` → `dist`
- Хостинг: Cloudflare Pages (проект `petfamily`, домен `petfamily.beliberde.net` через nginx)
- Бэкенд: FamilyOS API (`https://familyos.beliberde.net/api/v1`)

Деплой — автоматический: пуш в `main` → сборка и публикация на Pages.
