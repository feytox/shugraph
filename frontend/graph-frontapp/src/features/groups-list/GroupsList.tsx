import { Badge, Group, Text } from '@mantine/core';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import { Link, useSearchParams } from 'react-router';
import type { GroupItem } from '@/api/client';
import {
  accountsLabel,
  riskColors,
  riskLabels,
  schemeLabels,
  statusLabels,
} from '@/features/groups-list/groupLabels';
import { searchWithGroup, useSelectedGroupId } from '@/features/groups-list/useSelectedGroup';
import { formatRub } from '@/shared/money';
import classes from '@/features/groups-list/GroupsList.module.css';

dayjs.extend(utc);

function RiskBadge({ level, score }: { level: GroupItem['risk_level']; score: number }) {
  const description = `Риск ${riskLabels[level].toLowerCase()}: ${score}`;
  return (
    <Badge color={riskColors[level]} variant="filled" title={description} aria-label={description}>
      {score}
    </Badge>
  );
}

export function GroupsList({ items }: { items: GroupItem[] }) {
  const [params] = useSearchParams();
  const selectedId = useSelectedGroupId();
  return (
    <ul className={classes.list} aria-label="Список групп">
      {items.map((group) => (
        <li key={group.id}>
          <Link
            to={{ search: searchWithGroup(params, group.id) }}
            className={classes.row}
            aria-current={group.id === selectedId ? 'true' : undefined}
          >
            <Group justify="space-between" wrap="nowrap" mb="xs">
              <Text fw={600}>{schemeLabels[group.scheme]}</Text>
              <RiskBadge level={group.risk_level} score={group.risk_score} />
            </Group>
            <Text size="xs" c="dimmed">
              {accountsLabel(group.features.accounts_count)} ·{' '}
              {formatRub(group.features.total_amount)}
            </Text>
            <Group justify="space-between" wrap="nowrap">
              <Text size="xs" c="dimmed">
                {dayjs.utc(group.last_event_at).format('DD.MM.YYYY HH:mm')}
              </Text>
              <Text size="xs" c="dimmed">
                {statusLabels[group.status]}
              </Text>
            </Group>
          </Link>
        </li>
      ))}
    </ul>
  );
}
