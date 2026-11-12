import { Button, Center, Group, Loader, Stack, Text } from '@mantine/core';
import { GroupsList } from '@/features/groups-list/GroupsList';
import { useGroups } from '@/features/groups-list/useGroups';
import { ErrorNotice } from '@/shared/ErrorNotice';
import classes from '@/features/groups-list/GroupsPanel.module.css';

export function GroupsPanel() {
  const groups = useGroups();

  if (groups.isError && !groups.data) {
    return (
      <Center className={classes.fill} px="md">
        <ErrorNotice
          title="Не удалось загрузить список"
          details={groups.error.message}
          onRetry={() => groups.refetch()}
          retrying={groups.isFetching}
        />
      </Center>
    );
  }

  return (
    <Stack gap="sm" pb="md" className={classes.fill}>
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
      {(groups.isRefetchError || groups.isFetchNextPageError) && (
        <div className={classes.bottomNotice}>
          {groups.isRefetchError && (
            <ErrorNotice
              title="Не удалось обновить список"
              details={groups.error.message}
              onRetry={() => groups.refetch()}
              retrying={groups.isFetching}
            />
          )}
          {groups.isFetchNextPageError && (
            <ErrorNotice
              title="Не удалось загрузить следующую страницу"
              details={groups.error.message}
            />
          )}
        </div>
      )}
    </Stack>
  );
}
