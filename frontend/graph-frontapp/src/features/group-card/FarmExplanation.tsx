import { Stack, Table, Text, Title } from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import { useId } from 'react';
import type { Group } from '@/api/client';
import classes from './GroupCard.module.css';

export function FarmExplanation({
  group,
  risk,
}: {
  group: Group;
  risk: NonNullable<Group['risk_parts']['farm']>;
}) {
  const riskHeadingId = useId();
  const clientCount = new Set(group.logins.map(({ client_id }) => client_id)).size;
  const collectorSummary =
    risk.collector > 0
      ? 'Деньги с их счетов ушли на общий счёт для вывода — вероятный сборщик. Это признак фермы: возможно, за разными счетами стоит один организатор, который работает с одного устройства и собирает деньги на своём счёте.'
      : 'Общего счёта для вывода не видно, но несколько клиентов на одном устройстве — уже признак фермы: возможно, счетами управляет один человек.';
  const rows = [
    ['Тип схемы «ферма»', risk.scheme],
    [`Скорость: ${group.features.span_minutes} мин от первого до последнего события`, risk.speed],
    [`Клиентов на одном устройстве: ${clientCount}`, risk.clients],
    [
      risk.collector > 0 ? 'Есть общий счёт для вывода денег' : 'Общего счёта для вывода денег нет',
      risk.collector,
    ],
  ] as const;

  return (
    <>
      <Stack gap={4}>
        <Title order={3} size="h5">
          Что произошло
        </Title>
        <details>
          <Text component="summary" size="sm" className={classes.summary}>
            <span>
              С одного устройства заходили в личный кабинет разные клиенты (всего {clientCount}).
            </span>
            <IconChevronDown size={18} className={classes.chevron} aria-hidden="true" />
          </Text>
          <Text size="sm" mt={4}>
            {collectorSummary}
          </Text>
        </details>
      </Stack>
      <Stack gap={4}>
        <Title id={riskHeadingId} order={3} size="h5">
          Обоснование риска
        </Title>
        <Table aria-labelledby={riskHeadingId} fz="sm" horizontalSpacing="xs" withTableBorder>
          <Table.Thead bg="gray.0">
            <Table.Tr>
              <Table.Th scope="col">Признак</Table.Th>
              <Table.Th scope="col" ta="right">
                Баллы
              </Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {rows.map(([label, points]) => (
              <Table.Tr key={label}>
                <Table.Th scope="row" fw={400}>
                  {label}
                </Table.Th>
                <Table.Td ta="right" fw={600}>
                  +{points}
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      </Stack>
    </>
  );
}
