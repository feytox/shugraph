import { useRef } from 'react';
import { Alert, Button, Group, Loader, Stack, Text } from '@mantine/core';
import type { GroupGraph as GroupGraphData, GroupItem } from '@/api/client';
import { GraphLegend } from '@/features/group-graph/GraphLegend';
import { useCytoscapeGraph } from '@/features/group-graph/useCytoscapeGraph';
import { useGroupGraph } from '@/features/group-graph/useGroupGraph';
import classes from '@/features/group-graph/GroupGraph.module.css';

type Props = { groupId: string; scheme: GroupItem['scheme'] };

export function GroupGraph({ groupId, scheme }: Props) {
  const graph = useGroupGraph(groupId);

  if (graph.isPending) {
    return (
      <Group role="status" p="md">
        <Loader size="sm" />
        <Text>Загрузка графа…</Text>
      </Group>
    );
  }

  if (graph.isError) {
    return (
      <Alert color="red" title="Ошибка загрузки графа" m="md">
        <Stack gap="sm">
          <Text>{graph.error.message}</Text>
          <Button onClick={() => graph.refetch()} loading={graph.isFetching}>
            Повторить
          </Button>
        </Stack>
      </Alert>
    );
  }

  if (!graph.data) {
    return (
      <Text c="dimmed" p="md">
        Граф группы не найден
      </Text>
    );
  }

  return (
    <div className={classes.root}>
      <GraphCanvas scheme={scheme} graph={graph.data} />
      <GraphLegend />
    </div>
  );
}

function GraphCanvas({ scheme, graph }: { scheme: Props['scheme']; graph: GroupGraphData }) {
  const container = useRef<HTMLDivElement>(null);
  useCytoscapeGraph(container, scheme, graph);

  return (
    <div className={classes.canvasArea}>
      <div ref={container} className={classes.canvas} role="img" aria-label="Граф группы" />
    </div>
  );
}
