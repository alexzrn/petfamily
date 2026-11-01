#!/usr/bin/env node
// Валидатор точек карты Москвы. Читает JSON массив из файла.
// Печатает ошибки и выходит с кодом 1, если файл невалиден.
// Использование: node data/moscow/validate.mjs data/moscow/spots.json
import { readFileSync } from 'node:fs';

const CATEGORIES = new Set(['vet', 'boarding', 'grooming', 'park', 'shelter', 'shop']);
const REQUIRED = ['id', 'category', 'name', 'address', 'lat', 'lon', 'source', 'checked_at'];
const LAT_MIN = 55.1, LAT_MAX = 56.1, LON_MIN = 36.7, LON_MAX = 38.5;
const DATE_RE = /^\d{4}-\d{2}-\d{2}$/;
const URL_RE = /^https?:\/\/[^\s]+$/i;

function norm(s) {
  return String(s ?? '').trim().toLowerCase().replace(/\s+/g, ' ');
}

function main() {
  const file = process.argv[2];
  if (!file) {
    console.error('использование: node data/moscow/validate.mjs <файл.json>');
    process.exit(2);
  }
  let data;
  try {
    data = JSON.parse(readFileSync(file, 'utf8'));
  } catch (e) {
    console.error(`ошибка: файл не читается как JSON: ${e.message}`);
    process.exit(1);
  }
  if (!Array.isArray(data)) {
    console.error('ошибка: корень должен быть массивом точек');
    process.exit(1);
  }
  if (data.length === 0) {
    console.error('ошибка: файл не содержит ни одной точки');
    process.exit(1);
  }
  const errors = [];
  const ids = new Set();
  const addrName = new Set();
  data.forEach((spot, i) => {
    const tag = `точка #${i}${spot && spot.id ? ` (id=${spot.id})` : ''}`;
    if (!spot || typeof spot !== 'object') {
      errors.push(`${tag}: не объект`);
      return;
    }
    for (const f of REQUIRED) {
      if (spot[f] === undefined || spot[f] === null || String(spot[f]).trim() === '') {
        errors.push(`${tag}: отсутствует обязательное поле ${f}`);
      }
    }
    if (spot.category !== undefined && !CATEGORIES.has(spot.category)) {
      errors.push(`${tag}: категория ${JSON.stringify(spot.category)} вне списка vet/boarding/grooming/park/shelter/shop`);
    }
    if (typeof spot.lat === 'number' && (spot.lat < LAT_MIN || spot.lat > LAT_MAX)) {
      errors.push(`${tag}: широта ${spot.lat} вне Москвы (55.10-56.10)`);
    }
    if (typeof spot.lon === 'number' && (spot.lon < LON_MIN || spot.lon > LON_MAX)) {
      errors.push(`${tag}: долгота ${spot.lon} вне Москвы (36.70-38.50)`);
    }
    for (const b of ['open_24h', 'home_visit', 'inpatient']) {
      if (spot[b] !== undefined && spot[b] !== null && typeof spot[b] !== 'boolean') {
        errors.push(`${tag}: ${b} должен быть boolean`);
      }
    }
    if (spot.lat !== undefined && typeof spot.lat !== 'number') {
      errors.push(`${tag}: lat должен быть числом`);
    }
    if (spot.lon !== undefined && typeof spot.lon !== 'number') {
      errors.push(`${tag}: lon должен быть числом`);
    }
    if (spot.checked_at !== undefined && !DATE_RE.test(String(spot.checked_at))) {
      errors.push(`${tag}: checked_at должен быть датой YYYY-MM-DD`);
    }
    for (const u of ['source', 'site']) {
      if (spot[u] !== undefined && spot[u] !== null && String(spot[u]).trim() !== '' && !URL_RE.test(String(spot[u]).trim())) {
        errors.push(`${tag}: поле ${u} не похоже на URL`);
      }
    }
    if (spot.id !== undefined && String(spot.id).trim() !== '') {
      if (ids.has(spot.id)) errors.push(`${tag}: дубль id ${spot.id}`);
      ids.add(spot.id);
    }
    const key = `${norm(spot.address)}|${norm(spot.name)}`;
    if (spot.address !== undefined && spot.name !== undefined) {
      if (addrName.has(key)) errors.push(`${tag}: дубль по адресу и названию`);
      addrName.add(key);
    }
  });
  if (errors.length > 0) {
    for (const e of errors) console.error(`ошибка: ${e}`);
    process.exit(1);
  }
  console.log(`ok: точек ${data.length}, ошибок нет`);
}

main();
