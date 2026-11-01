#!/usr/bin/env node
// Импорт точек карты Москвы из CSV в JSON.
// Первая строка заголовок с именами полей схемы.
// Использование: node data/moscow/import-csv.mjs in.csv out.json
// Пустые необязательные поля (phone, hours, site, price) пишут как null.
import { readFileSync, writeFileSync } from 'node:fs';

const OPTIONAL = new Set(['phone', 'hours', 'site', 'price']);
const NUMERIC = new Set(['lat', 'lon']);

function parseCsv(text) {
  const rows = [];
  let row = [];
  let cell = '';
  let quoted = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (quoted) {
      if (c === '"') {
        if (text[i + 1] === '"') {
          cell += '"';
          i++;
        } else {
          quoted = false;
        }
      } else {
        cell += c;
      }
    } else if (c === '"') {
      quoted = true;
    } else if (c === ',') {
      row.push(cell);
      cell = '';
    } else if (c === '\n') {
      row.push(cell);
      rows.push(row);
      row = [];
      cell = '';
    } else if (c === '\r') {
      continue;
    } else {
      cell += c;
    }
  }
  if (cell !== '' || row.length > 0) {
    row.push(cell);
    rows.push(row);
  }
  return rows.filter((r) => r.some((v) => v.trim() !== ''));
}

function main() {
  const [input, output] = process.argv.slice(2);
  if (!input || !output) {
    console.error('использование: node data/moscow/import-csv.mjs <in.csv> <out.json>');
    process.exit(2);
  }
  const rows = parseCsv(readFileSync(input, 'utf8').replace(/^\uFEFF/, ''));
  if (rows.length < 2) {
    console.error('ошибка: в CSV нет строк данных');
    process.exit(1);
  }
  const header = rows[0].map((h) => h.trim());
  const spots = rows.slice(1).map((r) => {
    const spot = {};
    header.forEach((h, i) => {
      if (!h) return;
      const raw = (r[i] ?? '').trim();
      if (raw === '' && OPTIONAL.has(h)) {
        spot[h] = null;
        return;
      }
      spot[h] = NUMERIC.has(h) && raw !== '' ? Number(raw) : raw;
    });
    return spot;
  });
  writeFileSync(output, JSON.stringify(spots, null, 2) + '\n', 'utf8');
  console.log(`ok: строк ${spots.length} -> ${output}`);
}

main();
