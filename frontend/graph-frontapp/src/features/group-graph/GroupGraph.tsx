import { useEffect, useRef } from 'react';
import { Alert, Button, Group, Loader, Stack, Text, useMantineTheme } from '@mantine/core';
import cytoscape from 'cytoscape';
import type { GroupGraph as GroupGraphData, GroupItem } from '@/api/client';
import { GraphLegend } from '@/features/group-graph/GraphLegend';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { graphLayout } from '@/features/group-graph/graphLayout';
import { graphStylesheet } from '@/features/group-graph/graphStyle';
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

  return (
    <div className={classes.root}>
      <GraphCanvas scheme={scheme} graph={graph.data} />
      <GraphLegend />
    </div>
  );
}

function GraphCanvas({ scheme, graph }: { scheme: Props['scheme']; graph: GroupGraphData }) {
  const container = useRef<HTMLDivElement>(null);
  const theme = useMantineTheme();

  useEffect(() => {
    const cy = cytoscape({
      container: container.current,
      elements: toGraphElements(scheme, graph),
      style: graphStylesheet(theme),
    });
    cy.layout(graphLayout(scheme, cy)).run();
    return () => cy.destroy();
  }, [scheme, graph, theme]);

  return <div ref={container} className={classes.canvas} role="img" aria-label="Граф группы" />;
}
