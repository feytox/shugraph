import type { GroupItem } from '@/api/client';

export const schemeLabels: Record<GroupItem['scheme'], string> = {
  farm: 'Ферма',
  address: 'Адрес',
  chain: 'Цепочка',
  ring: 'Кольцо',
};

export const riskLabels: Record<GroupItem['risk_level'], string> = {
  low: 'Низкий',
  medium: 'Средний',
  high: 'Высокий',
};

export const riskColors: Record<GroupItem['risk_level'], string> = {
  low: 'green',
  medium: 'yellow',
  high: 'red',
};

export const statusLabels: Record<GroupItem['status'], string> = {
  new: 'Новая',
  approved: 'Подтверждена',
  declined: 'Отклонена',
};

const accountWords: Record<Intl.LDMLPluralRule, string> = {
  zero: 'счетов',
  one: 'счёт',
  two: 'счёта',
  few: 'счёта',
  many: 'счетов',
  other: 'счёта',
};
const russianPlural = new Intl.PluralRules('ru-RU');

export const accountsLabel = (count: number) =>
  `${count} ${accountWords[russianPlural.select(count)]}`;
