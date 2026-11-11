import { useMemo, useRef } from 'react';
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

export function GroupGraph({ groupId, scheme, lastEventAt }: Props) {
  const graph = useGroupGraph(groupId, lastEventAt);
  const elements = useMemo(
    () => (graph.data ? toGraphElements(scheme, graph.data) : undefined),
    [scheme, graph.data]
  );

  return (
    <div className={classes.root}>
      <div className={classes.canvasArea}>
        <GraphArea graph={graph} scheme={scheme} elements={elements} />
      </div>
      <GraphLegend elements={elements} />
    </div>
  );
}

function GraphArea({
  graph,
  scheme,
  elements,
}: {
  graph: GroupGraphQuery;
  scheme: GroupItem['scheme'];
  elements?: ElementsDefinition;
}) {
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
            <Button onClick={() => graph.refetch()} loading={graph.isFetching}>
              Повторить
            </Button>
          </Stack>
        </Alert>
      </Center>
    );
  }

  if (!elements) {
    return (
      <Center h="100%">
        <Text c="dimmed">Граф группы не найден</Text>
      </Center>
    );
  }

  return <GraphCanvas scheme={scheme} elements={elements} />;
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
