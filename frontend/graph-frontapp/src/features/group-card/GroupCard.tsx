import { Badge, Box, Group, Loader, Paper, Stack, Text } from '@mantine/core';
import type { ReactNode } from 'react';
import type { GroupItem } from '@/api/client';
import { FarmExplanation } from '@/features/group-card/FarmExplanation';
import { GroupNotFound, type useGroupDetails } from '@/features/group-card/useGroupDetails';
import { ErrorNotice } from '@/shared/ErrorNotice';
import { riskColors, riskLabels, statusColors, statusLabels } from '@/shared/groupLabels';
import { formatWholeRub } from '@/shared/money';
import classes from './GroupCard.module.css';

export function GroupCard({
  details,
  status,
  children,
}: {
  details: ReturnType<typeof useGroupDetails>;
  status: GroupItem['status'];
  children: ReactNode;
}) {
  const { data: group, isFetching, error, refetch } = details;

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

  return (
    <Stack px="md" pb="md" gap="md">
      {errorNotice}
      <Box>
        <Box aria-live="polite">
          {status !== 'new' && (
            <Badge color={statusColors[status]} variant="light" size="sm" tt="none" mb="xs">
              {statusLabels[status]}
            </Badge>
          )}
        </Box>
        <div className={classes.metrics}>
          <Paper
            component="section"
            aria-label={`Риск ${riskLabels[group.risk_level].toLowerCase()}: ${group.risk_score}`}
            className={classes.metric}
            bg={`${riskColors[group.risk_level]}.0`}
            radius="sm"
            p={8}
          >
            <Text size="xs" c="dimmed">
              Риск
            </Text>
            <Text size="md" fw={700} c={`${riskColors[group.risk_level]}.9`}>
              {group.risk_score}
            </Text>
            <Text size="xs" c={`${riskColors[group.risk_level]}.9`}>
              {riskLabels[group.risk_level]}
            </Text>
          </Paper>
          <Paper className={classes.metric} bg="gray.0" radius="sm" p={8}>
            <Text size="xs" c="dimmed">
              Счета
            </Text>
            <Text size="md" fw={700}>
              {accounts_count}
            </Text>
          </Paper>
          <Paper className={classes.metric} bg="gray.0" radius="sm" p={8}>
            <Text size="xs" c="dimmed">
              Оборот
            </Text>
            <Text size="md" fw={700}>
              {formatWholeRub(total_amount)}
            </Text>
          </Paper>
        </div>
      </Box>
      {farmRisk && <FarmExplanation group={group} risk={farmRisk} />}
      {children}
    </Stack>
  );
}
