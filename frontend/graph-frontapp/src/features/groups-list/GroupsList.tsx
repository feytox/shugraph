import { Badge, Group, Text } from '@mantine/core';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import type { GroupItem } from '@/api/client';
import {
  accountsLabel,
  riskColors,
  riskLabels,
  schemeLabels,
  statusLabels,
} from '@/features/groups-list/groupLabels';
import { formatRub } from '@/shared/money';
import classes from '@/features/groups-list/GroupsList.module.css';

dayjs.extend(utc);

export function GroupsList({ items }: { items: GroupItem[] }) {
  return (
    <ul className={classes.list} aria-label="Список групп">
      {items.map((group) => (
        <li key={group.id} className={classes.row}>
          <Group justify="space-between" wrap="nowrap">
            <Text fw={600}>{schemeLabels[group.scheme]}</Text>
            <Badge color={riskColors[group.risk_level]} variant="light">
              {riskLabels[group.risk_level]} · {group.risk_score}
            </Badge>
          </Group>
          <Text size="xs" c="dimmed">
            {accountsLabel(group.features.accounts_count)} ·{' '}
            {formatRub(group.features.total_amount)}
          </Text>
          <Group justify="space-between" wrap="nowrap">
            <Text size="xs" c="dimmed">
              {dayjs.utc(group.last_event_at).format('DD.MM.YYYY HH:mm:ss')} UTC
            </Text>
            <Text size="xs" c="dimmed">
              {statusLabels[group.status]}
            </Text>
          </Group>
        </li>
      ))}
    </ul>
  );
}
