import { List, Stack, Text, Title } from '@mantine/core';
import type { Group } from '@/api/client';
import { riskLabels } from '@/shared/groupLabels';

export function FarmExplanation({
  group,
  risk,
}: {
  group: Group;
  risk: NonNullable<Group['risk_parts']['farm']>;
}) {
  const clientCount = new Set(group.logins.map(({ client_id }) => client_id)).size;
  const collectorSummary =
    risk.collector > 0
      ? 'Деньги с их счетов ушли на общий счёт для вывода — вероятный сборщик. Это признак фермы: возможно, за разными счетами стоит один организатор, который работает с одного устройства и собирает деньги на своём счёте.'
      : 'Общего счёта для вывода не видно, но несколько клиентов на одном устройстве — уже признак фермы: возможно, счетами управляет один человек.';

  return (
    <>
      <Stack gap={4}>
        <Title order={3} size="h5">
          Что произошло:
        </Title>
        <Text size="sm">
          С одного устройства заходили в личный кабинет разные клиенты (всего {clientCount}).{' '}
          {collectorSummary}
        </Text>
      </Stack>
      <Stack gap={4}>
        <Title order={3} size="h5">
          Обоснование риска:
        </Title>
        <List size="sm" spacing={4} listStyleType="disc">
          <List.Item>Тип схемы «ферма»: +{risk.scheme}.</List.Item>
          <List.Item>
            Скорость: {group.features.span_minutes} мин от первого до последнего события — +
            {risk.speed}.
          </List.Item>
          <List.Item>
            Клиентов на одном устройстве: {clientCount} — +{risk.clients}.
          </List.Item>
          <List.Item>
            {risk.collector > 0
              ? `Есть общий счёт для вывода денег: +${risk.collector}.`
              : 'Общего счёта для вывода денег нет: +0.'}
          </List.Item>
        </List>
        <Text size="sm" fw={600}>
          Итого: {group.risk_score}/100 — {riskLabels[group.risk_level].toLowerCase()}.
        </Text>
      </Stack>
    </>
  );
}
