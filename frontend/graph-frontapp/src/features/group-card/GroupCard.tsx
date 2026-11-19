import { Badge, Box, Group, Loader, Stack, Text } from '@mantine/core';
import type { ReactNode } from 'react';
import { FarmExplanation } from '@/features/group-card/FarmExplanation';
import { GroupNotFound, useGroupDetails } from '@/features/group-card/useGroupDetails';
import { ErrorNotice } from '@/shared/ErrorNotice';
import { riskColors, riskLabels, statusColors, statusLabels } from '@/shared/groupLabels';
import { formatRub, formatWholeRub } from '@/shared/money';

export function GroupCard({
  groupId,
  lastEventAt,
  decision,
}: {
  groupId: string;
  lastEventAt: string;
  decision: (canDecide: boolean) => ReactNode;
}) {
  const { data: group, isFetching, error, refetch } = useGroupDetails(groupId, lastEventAt);

  if (error instanceof GroupNotFound) {
    return (
      <Text px="md" size="sm" c="dimmed">
        {error.message}
      </Text>
    );
  }

  const errorNotice = error && (
    <ErrorNotice
      title={error.message}
      wrapText
      onRetry={() => void refetch()}
      retrying={isFetching}
    />
  );
  if (!group && errorNotice) {
    return <Box px="md">{errorNotice}</Box>;
  }
  if (!group) {
    return (
      <Group gap="xs" px="md" role="status">
        <Loader size="sm" />
        <Text size="sm">Загрузка карточки…</Text>
      </Group>
    );
  }

  const { accounts_count, total_amount } = group.features;
  const farmRisk = group.scheme === 'farm' && group.logins.length > 0 && group.risk_parts.farm;
  const turnover = farmRisk ? formatWholeRub(total_amount) : formatRub(total_amount);

  return (
    <Stack px="md" pb="md" gap="md">
      {errorNotice}
      <Stack gap={4}>
        <Group justify="space-between" gap="xs">
          <Text size="sm" fw={700} c={`${riskColors[group.risk_level]}.9`}>
            Риск: {riskLabels[group.risk_level].toLowerCase()} ({group.risk_score}/100)
          </Text>
          <Badge
            color={statusColors[group.status]}
            variant="light"
            size="sm"
            tt="none"
            aria-live="polite"
          >
            {statusLabels[group.status]}
          </Badge>
        </Group>
        <Text size="sm">
          Счетов в группе: {accounts_count}, оборот: {turnover}.
        </Text>
      </Stack>
      {farmRisk && <FarmExplanation group={group} risk={farmRisk} />}
      {decision(group.status === 'new')}
    </Stack>
  );
}
