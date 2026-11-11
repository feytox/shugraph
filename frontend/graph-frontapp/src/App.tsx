import { Center, MantineProvider, Text } from '@mantine/core';
import { GroupGraph } from '@/features/group-graph/GroupGraph';
import { GroupsPanel } from '@/features/groups-list/GroupsPanel';
import {
  useDefaultGroupSelection,
  useSelectedGroup,
} from '@/features/groups-list/useSelectedGroup';
import { SidePanel } from '@/shared/SidePanel';
import classes from '@/App.module.css';

function AnalystScreen() {
  useDefaultGroupSelection();
  const { group: selected, notInList } = useSelectedGroup();
  return (
    <div className={classes.screen}>
      <SidePanel side="left" title="Подозрительные группы" storageKey="groups-panel">
        <GroupsPanel />
      </SidePanel>
      <main className={classes.graph}>
        {selected && (
          <GroupGraph
            key={selected.id}
            groupId={selected.id}
            scheme={selected.scheme}
            lastEventAt={selected.last_event_at}
          />
        )}
        {notInList && (
          <Center h="100%">
            <Text c="dimmed">Группа не найдена в загруженном списке</Text>
          </Center>
        )}
      </main>
      <aside className={classes.card} aria-label="Карточка группы" />
    </div>
  );
}

export function App() {
  return (
    <MantineProvider>
      <AnalystScreen />
    </MantineProvider>
  );
}
