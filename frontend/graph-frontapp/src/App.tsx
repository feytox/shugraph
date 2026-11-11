import {
  Badge,
  Container,
  MantineProvider,
  Stack,
  Table,
  Title,
  Text,
  Loader,
  Group,
  Alert,
  Button,
} from '@mantine/core';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import type { GroupItem } from '@/api/client';
import { useGroups } from '@/features/groups-list/useGroups';
import { formatRub } from '@/shared/money';

dayjs.extend(utc);
const schemes = { farm: 'Ферма', address: 'Адрес', chain: 'Цепочка', ring: 'Кольцо' };
const risks = { low: 'Низкий', medium: 'Средний', high: 'Высокий' };
const riskColors = { low: 'green', medium: 'yellow', high: 'red' };
const statuses = { new: 'Новая', approved: 'Подтверждена', declined: 'Отклонена' };

function GroupsList({ items }: { items: GroupItem[] }) {
  return (
    <Table striped withTableBorder>
      <Table.Thead>
        <Table.Tr>
          {['Тип схемы', 'Риск', 'Число счетов', 'Сумма', 'Последнее событие', 'Статус'].map(
            (label) => (
              <Table.Th key={label}>{label}</Table.Th>
            )
          )}
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {items.map((group) => (
          <Table.Tr key={group.id}>
            <Table.Td>{schemes[group.scheme]}</Table.Td>
            <Table.Td>
              <Badge color={riskColors[group.risk_level]} variant="light">
                {risks[group.risk_level]} · {group.risk_score}
              </Badge>
            </Table.Td>
            <Table.Td>{group.features.accounts_count}</Table.Td>
            <Table.Td>{formatRub(group.features.total_amount)}</Table.Td>
            <Table.Td>{dayjs.utc(group.last_event_at).format('DD.MM.YYYY HH:mm:ss')} UTC</Table.Td>
            <Table.Td>{statuses[group.status]}</Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  );
}

function GroupsScreen() {
  const groups = useGroups();
  return (
    <Container fluid p="xl">
      <Stack>
        <Title order={1}>Подозрительные группы</Title>
        {groups.isError && !groups.data && (
          <Alert color="red" title="Ошибка загрузки">
            <Stack gap="sm">
              <Text>{groups.error.message}</Text>
              <Button onClick={() => groups.refetch()} loading={groups.isFetching}>
                Повторить
              </Button>
            </Stack>
          </Alert>
        )}
        {groups.isRefetchError && groups.data && (
          <Alert color="red" title="Не удалось обновить список">
            {groups.error.message}
          </Alert>
        )}
        {groups.isPending && (
          <Group role="status">
            <Loader size="sm" />
            <Text>Загрузка групп…</Text>
          </Group>
        )}
        {groups.data &&
          (groups.data.items.length ? (
            <GroupsList items={groups.data.items} />
          ) : (
            <Text>
              Подозрительных групп пока нет. Запустите сценарий через Swagger graph-admin-api
            </Text>
          ))}
        {groups.isFetchNextPageError && (
          <Alert color="red" title="Не удалось загрузить следующую страницу">
            {groups.error.message}
          </Alert>
        )}
        {groups.hasNextPage && (
          <Button onClick={() => groups.fetchNextPage()} loading={groups.isFetchingNextPage}>
            {groups.isFetchNextPageError ? 'Повторить загрузку' : 'Показать ещё'}
          </Button>
        )}
      </Stack>
    </Container>
  );
}

export function App() {
  return (
    <MantineProvider>
      <GroupsScreen />
    </MantineProvider>
  );
}
