import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';

dayjs.extend(utc);

export const formatTimeUtc = (isoDate: string) => dayjs.utc(isoDate).format('HH:mm');

export const formatDateTimeUtc = (isoDate: string) => dayjs.utc(isoDate).format('DD.MM.YY HH:mm');

const dayUtc = (isoDate: string) => dayjs.utc(isoDate).format('DD.MM.YY');

export const formatDecidedAtUtc = (isoDate: string) =>
  `${dayUtc(isoDate)}, ${formatTimeUtc(isoDate)} UTC`;

export const formatEventRangeUtc = (firstIso: string, lastIso: string) =>
  dayUtc(firstIso) === dayUtc(lastIso)
    ? `${dayUtc(firstIso)}, ${formatTimeUtc(firstIso)}–${formatTimeUtc(lastIso)} UTC`
    : `${dayUtc(firstIso)}, ${formatTimeUtc(firstIso)} – ${dayUtc(lastIso)}, ${formatTimeUtc(lastIso)} UTC`;
