import { MantineProvider } from '@mantine/core';
import { GroupCard } from '@/features/group-card/GroupCard';
import { GroupDecision } from '@/features/group-decision/GroupDecision';
import { GraphFrame, GraphMessage, GroupGraph } from '@/features/group-graph/GroupGraph';
import { GroupsPanel, GroupsPanelError } from '@/features/groups-list/GroupsPanel';
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
      <SidePanel
        side="left"
        title="Подозрительные группы"
        storageKey="groups-panel"
        footer={<GroupsPanelError />}
      >
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
          <GraphFrame>
            <GraphMessage>Группа не найдена в загруженном списке</GraphMessage>
          </GraphFrame>
        )}
      </main>
      {selected && (
        <SidePanel side="right" title="Карточка расследования" storageKey="investigation-panel">
          <GroupCard
            key={selected.id}
            groupId={selected.id}
            decision={<GroupDecision groupId={selected.id} />}
          />
        </SidePanel>
      )}
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
