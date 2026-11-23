import { Badge, Box, Group, Loader, Stack, Text } from '@mantine/core';
import type { ReactNode } from 'react';
import type { GroupItem } from '@/api/client';
import { FarmExplanation } from '@/features/group-card/FarmExplanation';
import { GroupNotFound, type useGroupDetails } from '@/features/group-card/useGroupDetails';
import { ErrorNotice } from '@/shared/ErrorNotice';
import { riskColors, riskLabels, statusColors, statusLabels } from '@/shared/groupLabels';
import { formatWholeRub } from '@/shared/money';

export function GroupCard({
  details,
  listedStatus,
  decision,
}: {
  details: ReturnType<typeof useGroupDetails>;
  listedStatus: GroupItem['status'] | undefined;
  decision: (state: { isOpen: boolean; evidenceReady: boolean }) => ReactNode;
}) {
  const { data: group, dataUpdatedAt, isFetching, fetchStatus, error, refetch } = details;

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
  const status = group.status === 'new' ? (listedStatus ?? group.status) : group.status;
  const farmRisk = group.scheme === 'farm' && group.logins.length > 0 && group.risk_parts.farm;

  return (
    <Stack px="md" pb="md" gap="md">
      {errorNotice}
      <Stack gap={4}>
        <Group justify="space-between" gap="xs">
          <Text size="sm" fw={700} c={`${riskColors[group.risk_level]}.9`}>
            Риск: {riskLabels[group.risk_level].toLowerCase()} ({group.risk_score}/100)
          </Text>
          <Badge
            color={statusColors[status]}
            variant="light"
            size="sm"
            tt="none"
            aria-live="polite"
          >
            {statusLabels[status]}
          </Badge>
        </Group>
        <Text size="sm">
          Счетов в группе: {accounts_count}, оборот: {formatWholeRub(total_amount)}.
        </Text>
      </Stack>
      {farmRisk && <FarmExplanation group={group} risk={farmRisk} />}
      {decision({
        isOpen: status === 'new',
        evidenceReady: dataUpdatedAt > 0 && fetchStatus === 'idle' && !error,
      })}
    </Stack>
  );
}
