import { Group, Text } from '@mantine/core';
import { Link, useSearchParams } from 'react-router';
import type { GroupItem } from '@/api/client';
import { RiskBadge, StatusBadge } from '@/shared/GroupBadges';
import { accountsLabel, schemeLabels } from '@/shared/groupLabels';
import { searchWithGroup, useSelectedGroupId } from '@/features/groups-list/useSelectedGroup';
import { formatDateTimeUtc } from '@/shared/dates';
import { formatRub } from '@/shared/money';
import classes from '@/features/groups-list/GroupsList.module.css';

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
