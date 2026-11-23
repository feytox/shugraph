import { MantineProvider } from '@mantine/core';
import { GroupCard } from '@/features/group-card/GroupCard';
import { GroupNotFound, useGroupDetails } from '@/features/group-card/useGroupDetails';
import { GroupDecision } from '@/features/group-decision/GroupDecision';
import { GraphFrame, GraphMessage, GroupGraph } from '@/features/group-graph/GroupGraph';
import { GroupsPanel, GroupsPanelError } from '@/features/groups-list/GroupsPanel';
import {
  useDefaultGroupSelection,
  useGroupSummary,
  useSelectedGroupId,
} from '@/features/groups-list/useSelectedGroup';
import { SidePanel } from '@/shared/SidePanel';
import classes from '@/App.module.css';

function AnalystScreen() {
  useDefaultGroupSelection();
  const selectedId = useSelectedGroupId();
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
      {selectedId ? (
        <SelectedGroup key={selectedId} groupId={selectedId} />
      ) : (
        <main className={classes.graph} />
      )}
    </div>
  );
}

function SelectedGroup({ groupId }: { groupId: string }) {
  const summary = useGroupSummary(groupId);
  const details = useGroupDetails(groupId, summary);
  const notFound = details.error instanceof GroupNotFound;
  const group = notFound ? undefined : (details.data ?? summary);

  return (
    <>
      <main className={classes.graph}>
        {group ? (
          <GroupGraph
            groupId={groupId}
            scheme={group.scheme}
            lastEventAt={summary?.last_event_at ?? group.last_event_at}
          />
        ) : (
          <GraphFrame>
            <GraphMessage>
              {notFound
                ? 'Группа не найдена'
                : details.isError
                  ? 'Не удалось загрузить группу'
                  : 'Загрузка группы…'}
            </GraphMessage>
          </GraphFrame>
        )}
      </main>
      <SidePanel side="right" title="Карточка расследования" storageKey="investigation-panel">
        <GroupCard
          details={details}
          listedStatus={summary?.status}
          decision={(state) => <GroupDecision groupId={groupId} {...state} />}
        />
      </SidePanel>
    </>
  );
}

export function App() {
  return (
    <MantineProvider>
      <AnalystScreen />
    </MantineProvider>
  );
}
