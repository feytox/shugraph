import { useMemo, useRef, useState, type ReactNode } from 'react';
import { Alert, Button, Center, Group, Loader, Stack, Text } from '@mantine/core';
import type { ElementsDefinition } from 'cytoscape';
import type { GroupItem } from '@/api/client';
import { GraphLegend } from '@/features/group-graph/GraphLegend';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { useCytoscapeGraph } from '@/features/group-graph/useCytoscapeGraph';
import { useGroupGraph } from '@/features/group-graph/useGroupGraph';
import classes from '@/features/group-graph/GroupGraph.module.css';

type Props = { groupId: string; scheme: GroupItem['scheme']; lastEventAt: string };
type GroupGraphQuery = ReturnType<typeof useGroupGraph>;

export function GraphFrame({
  elements,
  children,
}: {
  elements?: ElementsDefinition;
  children: ReactNode;
}) {
  return (
    <div className={classes.root}>
      <div className={classes.canvasArea}>{children}</div>
      <GraphLegend elements={elements} />
    </div>
  );
}

export function GraphMessage({ children }: { children: ReactNode }) {
  return (
    <Center h="100%">
      <Text c="dimmed">{children}</Text>
    </Center>
  );
}

export function GroupGraph({ groupId, scheme, lastEventAt }: Props) {
  const graph = useGroupGraph(groupId, lastEventAt);
  const fetched = useMemo(
    () => (graph.data ? toGraphElements(scheme, graph.data) : undefined),
    [scheme, graph.data]
  );
  const [drawn, setDrawn] = useState(fetched);
  if (fetched && fetched !== drawn) setDrawn(fetched);

  return (
    <GraphFrame elements={drawn}>
      {drawn ? (
        <>
          <GraphCanvas scheme={scheme} elements={drawn} />
          {graph.isError && <RefreshError graph={graph} />}
        </>
      ) : (
        <GraphStatus graph={graph} />
      )}
    </GraphFrame>
  );
}

function GraphStatus({ graph }: { graph: GroupGraphQuery }) {
  if (graph.isPending) {
    return (
      <Center h="100%">
        <Group role="status">
          <Loader size="sm" />
          <Text>Загрузка графа…</Text>
        </Group>
      </Center>
    );
  }

  if (graph.isError) {
    return (
      <Center h="100%">
        <Alert color="red" title="Ошибка загрузки графа">
          <Stack gap="sm">
            <Text>{graph.error.message}</Text>
            <RetryButton graph={graph} />
          </Stack>
        </Alert>
      </Center>
    );
  }

  return <GraphMessage>Граф группы не найден</GraphMessage>;
}

function RefreshError({ graph }: { graph: GroupGraphQuery }) {
  return (
    <Alert color="red" title="Не удалось обновить граф" className={classes.refreshError}>
      <Stack gap="xs">
        <Text size="sm">На экране последняя загруженная версия.</Text>
        <RetryButton graph={graph} />
      </Stack>
    </Alert>
  );
}

function RetryButton({ graph }: { graph: GroupGraphQuery }) {
  return (
    <Button onClick={() => graph.refetch()} loading={graph.isFetching}>
      Повторить
    </Button>
  );
}

function GraphCanvas({
  scheme,
  elements,
}: {
  scheme: GroupItem['scheme'];
  elements: ElementsDefinition;
}) {
  const container = useRef<HTMLDivElement>(null);
  useCytoscapeGraph(container, scheme, elements);

  return <div ref={container} className={classes.canvas} role="img" aria-label="Граф группы" />;
}
