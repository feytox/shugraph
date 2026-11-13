import { Button, Group, Loader, Stack, Text } from '@mantine/core';
import { GroupsList } from '@/features/groups-list/GroupsList';
import { useGroups } from '@/features/groups-list/useGroups';
import { ErrorNotice } from '@/shared/ErrorNotice';
import classes from '@/features/groups-list/GroupsPanel.module.css';

type GroupsQuery = ReturnType<typeof useGroups>;

export function GroupsPanel() {
  const groups = useGroups();

  return (
    <Stack gap="sm" pb="md">
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
      {groups.hasNextPage && <NextPageSlot groups={groups} />}
    </Stack>
  );
}

function NextPageSlot({ groups }: { groups: GroupsQuery }) {
  return (
    <div className={classes.nextPageSlot}>
      {groups.isFetchNextPageError ? (
        <ErrorNotice
          title="Не удалось загрузить ещё"
          onRetry={() => groups.fetchNextPage()}
          retrying={groups.isFetchingNextPage}
        />
      ) : (
        <Button
          variant="default"
          onClick={() => groups.fetchNextPage()}
          loading={groups.isFetchingNextPage}
        >
          Показать ещё
        </Button>
      )}
    </div>
  );
}

export function GroupsPanelError() {
  const groups = useGroups();
  if (!hasWholeListError(groups)) return null;

  return (
    <ErrorNotice
      title={groups.isRefetchError ? 'Не удалось обновить список' : 'Не удалось загрузить список'}
      onRetry={() => groups.refetch()}
      retrying={groups.isFetching}
    />
  );
}

function hasWholeListError(groups: GroupsQuery) {
  return groups.isLoadingError || groups.isRefetchError;
}
