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
import { theme } from '@/theme';
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
  const { summary, listError } = useGroupSummary(groupId);
  const currentSummary = listError ? undefined : summary;
  const details = useGroupDetails(groupId, currentSummary);
  const notFound = details.error instanceof GroupNotFound;
  const group = notFound ? undefined : (details.data ?? summary);
  const detailStatus = details.data?.status;
  const status = detailStatus && detailStatus !== 'new' ? detailStatus : (summary?.status ?? 'new');
  const evidenceReady =
    details.dataUpdatedAt > 0 && details.fetchStatus === 'idle' && !details.error;

  return (
    <>
      <main className={classes.graph}>
        {group ? (
          <GroupGraph
            groupId={groupId}
            scheme={group.scheme}
            lastEventAt={currentSummary?.last_event_at ?? group.last_event_at}
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
        <GroupCard details={details} status={status}>
          <GroupDecision
            groupId={groupId}
            isOpen={status === 'new'}
            evidenceReady={evidenceReady}
          />
        </GroupCard>
      </SidePanel>
    </>
  );
}

export function App() {
  return (
    <MantineProvider theme={theme}>
      <AnalystScreen />
    </MantineProvider>
  );
}
