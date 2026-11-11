import { MantineProvider } from '@mantine/core';
import { GroupGraph } from '@/features/group-graph/GroupGraph';
import { GroupsPanel } from '@/features/groups-list/GroupsPanel';
import { useSelectedGroup } from '@/features/groups-list/useSelectedGroup';
import classes from '@/App.module.css';

function AnalystScreen() {
  const selected = useSelectedGroup();
  return (
    <div className={classes.screen}>
      <aside className={classes.groups} aria-label="Группы">
        <GroupsPanel />
      </aside>
      <main className={classes.graph}>
        {selected && (
          <GroupGraph key={selected.id} groupId={selected.id} scheme={selected.scheme} />
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
