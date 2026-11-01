import { defineConfig } from 'astro/config';

// build.format: 'file' — страницы собираются как /map.html, /spot/vet-001.html.
// Cloudflare Pages отдаёт их по адресам /map и /spot/vet-001 без редиректа на слеш.
export default defineConfig({
  build: { format: 'file' },
  trailingSlash: 'never',
});
