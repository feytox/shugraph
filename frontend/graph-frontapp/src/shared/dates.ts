import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';

dayjs.extend(utc);

export const formatTimeUtc = (isoDate: string) => dayjs.utc(isoDate).format('HH:mm');

export const formatDateTimeUtc = (isoDate: string) => dayjs.utc(isoDate).format('DD.MM.YY HH:mm');
