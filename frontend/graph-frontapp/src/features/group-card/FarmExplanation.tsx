import { Text } from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import type { Group } from '@/api/client';
import { farmClientCount } from './farmClients';
import classes from './GroupCard.module.css';

type FarmRisk = NonNullable<Group['risk_parts']['farm']>;

export function FarmSummary({ group, risk }: { group: Group; risk: FarmRisk }) {
  const collectorSummary =
    risk.collector > 0
      ? 'Деньги с их счетов ушли на общий счёт для вывода — вероятный сборщик. Это признак фермы: возможно, за разными счетами стоит один организатор, который работает с одного устройства и собирает деньги на своём счёте.'
      : 'Общего счёта для вывода не видно, но несколько клиентов на одном устройстве — уже признак фермы: возможно, счетами управляет один человек.';

  return (
    <details>
      <Text component="summary" size="sm" className={classes.summary}>
        <span>
          С одного устройства заходили в личный кабинет разные клиенты (всего{' '}
          {farmClientCount(group)}).
        </span>
        <IconChevronDown size={18} className={classes.chevron} aria-hidden="true" />
      </Text>
      <Text size="sm" mt={4}>
        {collectorSummary}
      </Text>
    </details>
  );
}

export function FarmRiskParts({ risk, titleId }: { risk: FarmRisk; titleId: string }) {
  const rows = [
    ['Схема «ферма»', risk.scheme],
    ['Скорость', risk.speed],
    ['Клиентов на устройстве', risk.clients],
    ['Общий счёт для вывода', risk.collector],
  ] as const;

  return (
    <table aria-labelledby={titleId} className={classes.parts}>
      <tbody>
        {rows.map(([label, points]) => (
          <tr key={label} className={points === 0 ? classes.zero : undefined}>
            <td>{label}</td>
            <td>+{points}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
