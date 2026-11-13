import { Badge, Group, Text } from '@mantine/core';
import { Link, useSearchParams } from 'react-router';
import type { GroupItem } from '@/api/client';
import {
  accountsLabel,
  riskColors,
  riskLabels,
  schemeLabels,
  statusColors,
  statusLabels,
} from '@/features/groups-list/groupLabels';
import { searchWithGroup, useSelectedGroupId } from '@/features/groups-list/useSelectedGroup';
import { formatDateTimeUtc } from '@/shared/dates';
import { formatRub } from '@/shared/money';
import classes from '@/features/groups-list/GroupsList.module.css';

function RiskBadge({ level, score }: { level: GroupItem['risk_level']; score: number }) {
  return (
    <Badge
      color={riskColors[level]}
      variant="filled"
      size="lg"
      fw={800}
      className={classes.keepWidth}
      aria-label={`Риск ${riskLabels[level].toLowerCase()}: ${score}`}
    >
      {score}
    </Badge>
  );
}

function StatusBadge({ status }: { status: GroupItem['status'] }) {
  return (
    <Badge
      color={statusColors[status]}
      variant="light"
      size="sm"
      tt="none"
      className={classes.keepWidth}
    >
      {statusLabels[status]}
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
              <Group gap="xs" wrap="nowrap">
                <Text fw={600}>{schemeLabels[group.scheme]}</Text>
                <StatusBadge status={group.status} />
              </Group>
              <RiskBadge level={group.risk_level} score={group.risk_score} />
            </Group>
            <Group justify="space-between" wrap="nowrap">
              <Text size="xs" c="dimmed" truncate>
                {accountsLabel(group.features.accounts_count)} ·{' '}
                {formatRub(group.features.total_amount)}
              </Text>
              <Text size="xs" c="dimmed" className={classes.keepWidth}>
                {formatDateTimeUtc(group.last_event_at)}
              </Text>
            </Group>
          </Link>
        </li>
      ))}
    </ul>
  );
}
