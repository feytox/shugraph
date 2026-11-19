const rubles = (fractionDigits: number) =>
  new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });

const wholeRubles = rubles(0);
const withKopecks = rubles(2);

export const formatWholeRub = (kopecks: number) => wholeRubles.format(kopecks / 100);

export const formatRub = (kopecks: number) =>
  kopecks % 100 === 0 ? formatWholeRub(kopecks) : withKopecks.format(kopecks / 100);
