import { Badge, Box, Group, Loader, Stack, Text } from '@mantine/core';
import type { ReactNode } from 'react';
import { FarmExplanation } from '@/features/group-card/FarmExplanation';
import { useGroupDetails } from '@/features/group-card/useGroupDetails';
import { ErrorNotice } from '@/shared/ErrorNotice';
import { riskColors, riskLabels, statusColors, statusLabels } from '@/shared/groupLabels';
import { formatRub, formatWholeRub } from '@/shared/money';

export function GroupCard({ groupId, decision }: { groupId: string; decision: ReactNode }) {
  const { data: group, isPending, isFetching, isError, refetch } = useGroupDetails(groupId);

  if (isPending) {
    return (
      <Group gap="xs" px="md" role="status">
        <Loader size="sm" />
        <Text size="sm">Загрузка карточки…</Text>
      </Group>
    );
  }
  if (!group) {
    return (
      <Box px="md">
        {isError ? (
          <ErrorNotice
            title="Не удалось загрузить карточку группы"
            onRetry={() => void refetch()}
            retrying={isFetching}
          />
        ) : (
          <Text size="sm" c="dimmed">
            Карточка группы не найдена
          </Text>
        )}
      </Box>
    );
  }

  const { accounts_count, total_amount } = group.features;
  const turnover = group.scheme === 'farm' ? formatWholeRub(total_amount) : formatRub(total_amount);

  return (
    <Stack px="md" pb="md" gap="md">
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
      <FarmExplanation group={group} />
      {group.status === 'new' && decision}
    </Stack>
  );
}
