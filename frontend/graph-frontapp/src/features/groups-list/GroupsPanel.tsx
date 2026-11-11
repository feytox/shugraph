import { Alert, Button, Group, Loader, Stack, Text } from '@mantine/core';
import { GroupsList } from '@/features/groups-list/GroupsList';
import { useGroups } from '@/features/groups-list/useGroups';

export function GroupsPanel() {
  const groups = useGroups();
  return (
    <Stack gap="sm" pb="md">
      {groups.isError && !groups.data && (
        <Alert color="red" title="Ошибка загрузки" mx="md">
          <Stack gap="sm">
            <Text>{groups.error.message}</Text>
            <Button onClick={() => groups.refetch()} loading={groups.isFetching}>
              Повторить
            </Button>
          </Stack>
        </Alert>
      )}
      {groups.isRefetchError && groups.data && (
        <Alert color="red" title="Не удалось обновить список" mx="md">
          {groups.error.message}
        </Alert>
      )}
      {groups.isPending && (
        <Group role="status" px="md">
          <Loader size="sm" />
          <Text>Загрузка групп…</Text>
        </Group>
      )}
      {groups.data &&
        (groups.data.items.length ? (
          <GroupsList items={groups.data.items} />
        ) : (
          <Text px="md">
            Подозрительных групп пока нет. Запустите сценарий через Swagger graph-admin-api
          </Text>
        ))}
      {groups.isFetchNextPageError && (
        <Alert color="red" title="Не удалось загрузить следующую страницу" mx="md">
          {groups.error.message}
        </Alert>
      )}
      {groups.hasNextPage && (
        <Button
          variant="default"
          onClick={() => groups.fetchNextPage()}
          loading={groups.isFetchingNextPage}
          mx="md"
        >
          {groups.isFetchNextPageError ? 'Повторить загрузку' : 'Показать ещё'}
        </Button>
      )}
    </Stack>
  );
}
