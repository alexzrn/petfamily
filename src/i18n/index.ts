import ru, { type Dictionary } from './ru';

// en и th добавляются сюда же: import en from './en'; locales.en = en.
// Вёрстка берёт тексты только из этого словаря.
export const locales: Record<string, Dictionary> = { ru };
export const defaultLocale = 'ru';

export function getDictionary(locale: string = defaultLocale): Dictionary {
  return locales[locale] ?? locales[defaultLocale];
}
