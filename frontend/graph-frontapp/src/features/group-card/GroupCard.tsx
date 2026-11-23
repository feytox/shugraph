import { Box, Group, Loader, Stack, Text } from '@mantine/core';
import { useId, type ReactNode } from 'react';
import type { GroupItem } from '@/api/client';
import { FarmRiskParts, FarmSummary } from '@/features/group-card/FarmExplanation';
import { farmClientCount } from '@/features/group-card/farmClients';
import { GroupNotFound, type useGroupDetails } from '@/features/group-card/useGroupDetails';
import { ErrorNotice } from '@/shared/ErrorNotice';
import { RiskBadge, StatusBadge } from '@/shared/GroupBadges';
import { formatDecidedAtUtc, formatEventRangeUtc } from '@/shared/dates';
import { schemeLabels } from '@/shared/groupLabels';
import { formatWholeRub } from '@/shared/money';
import classes from './GroupCard.module.css';

function Fact({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className={classes.fact}>
      <small>{label}</small>
      <span>{value}</span>
    </div>
  );
}

export function GroupCard({
  details,
  status,
  children,
}: {
  details: ReturnType<typeof useGroupDetails>;
  status: GroupItem['status'];
  children: ReactNode;
}) {
  const riskTitleId = useId();
  const decisionTitleId = useId();
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

  const { accounts_count, span_minutes, total_amount } = group.features;
  const farmRisk = group.scheme === 'farm' && group.logins.length > 0 && group.risk_parts.farm;
  const decision = status !== 'new' && group.decision;

  return (
    <>
      <Stack px="md" pb="sm" gap="md">
        {errorNotice}
        <Group justify="space-between" gap="sm" wrap="nowrap">
          <Text fz={20} fw={700}>
            {schemeLabels[group.scheme]}
          </Text>
          <RiskBadge level={group.risk_level} score={group.risk_score} />
        </Group>
        {farmRisk && <FarmSummary group={group} risk={farmRisk} />}
        <div className={classes.facts}>
          <Fact label="счетов" value={accounts_count} />
          {farmRisk && <Fact label="клиентов" value={farmClientCount(group)} />}
          <Fact label="за" value={`${span_minutes} мин`} />
          <Fact label="оборот" value={formatWholeRub(total_amount)} />
        </div>
        {farmRisk && (
          <div>
            <p id={riskTitleId} className={classes.sectionTitle}>
              Из чего риск
            </p>
            <FarmRiskParts risk={farmRisk} titleId={riskTitleId} />
          </div>
        )}
        <section aria-labelledby={decisionTitleId} className={classes.decision}>
          <div className={classes.decisionHead}>
            <p id={decisionTitleId} className={classes.sectionTitle}>
              Решение
            </p>
            <Box aria-live="polite">{status !== 'new' && <StatusBadge status={status} />}</Box>
          </div>
          {decision && (
            <>
              <p className={classes.comment}>{decision.comment}</p>
              <span className={classes.decided}>{formatDecidedAtUtc(decision.decided_at)}</span>
            </>
          )}
          {children}
        </section>
      </Stack>
      <p className={classes.dates}>
        события {formatEventRangeUtc(group.first_event_at, group.last_event_at)}
      </p>
    </>
  );
}
