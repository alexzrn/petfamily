// Данные точек карты и общие подписи. Единственный источник — data/moscow/spots.json.
import raw from '../../data/moscow/spots.json';

export type Category = 'vet' | 'shelter' | 'park' | 'boarding' | 'grooming' | 'shop';

export type Spot = {
  id: string;
  category: Category;
  name: string;
  address: string;
  district: string | null;
  phone: string | null;
  hours: string | null;
  site: string | null;
  price: string | null;
  lat: number;
  lon: number;
  source: string;
  checked_at: string;
  open_24h: boolean | null;
  home_visit: boolean | null;
  inpatient: boolean | null;
  geo?: string;
};

export const CATEGORY_LABELS: Record<Category, string> = {
  vet: 'Клиники',
  park: 'Площадки',
  shelter: 'Приюты',
  boarding: 'Передержка',
  grooming: 'Груминг',
  shop: 'Магазины',
};

export const CATEGORY_ORDER: Category[] = ['vet', 'park', 'shelter', 'boarding', 'grooming', 'shop'];

export const spots = raw as Spot[];

export const counts = CATEGORY_ORDER.reduce(
  (acc, cat) => {
    acc[cat] = spots.filter((s) => s.category === cat).length;
    return acc;
  },
  {} as Record<Category, number>,
);

export const total = spots.length;

/** Телефон в виде, пригодном для tel: */
export function telHref(phone: string): string {
  const digits = phone.replace(/[^\d+]/g, '');
  if (digits.startsWith('8')) return 'tel:+7' + digits.slice(1);
  if (digits.startsWith('+')) return 'tel:' + digits;
  return 'tel:+7' + digits;
}

export function spotUrl(id: string): string {
  return '/spot/' + id;
}
