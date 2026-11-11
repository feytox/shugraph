import { useRef } from 'react';
import { Alert, Button, Center, Group, Loader, Stack, Text } from '@mantine/core';
import type { GroupGraph as GroupGraphData, GroupItem } from '@/api/client';
import { GraphLegend } from '@/features/group-graph/GraphLegend';
import { useCytoscapeGraph } from '@/features/group-graph/useCytoscapeGraph';
import { useGroupGraph } from '@/features/group-graph/useGroupGraph';
import classes from '@/features/group-graph/GroupGraph.module.css';

type Props = { groupId: string; scheme: GroupItem['scheme'] };

export function GroupGraph({ groupId, scheme }: Props) {
  return (
    <div className={classes.root}>
      <div className={classes.canvasArea}>
        <GraphArea groupId={groupId} scheme={scheme} />
      </div>
      <GraphLegend />
    </div>
  );
}

function GraphArea({ groupId, scheme }: Props) {
  const graph = useGroupGraph(groupId);

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

  if (!graph.data) {
    return (
      <Center h="100%">
        <Text c="dimmed">Граф группы не найден</Text>
      </Center>
    );
  }

  return <GraphCanvas scheme={scheme} graph={graph.data} />;
}

function GraphCanvas({ scheme, graph }: { scheme: Props['scheme']; graph: GroupGraphData }) {
  const container = useRef<HTMLDivElement>(null);
  useCytoscapeGraph(container, scheme, graph);

  return <div ref={container} className={classes.canvas} role="img" aria-label="Граф группы" />;
}
