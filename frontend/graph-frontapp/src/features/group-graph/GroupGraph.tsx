import { useEffect, useRef } from 'react';
import { Alert, Button, Group, Loader, Stack, Text, useMantineTheme } from '@mantine/core';
import cytoscape, { type Core } from 'cytoscape';
import type { GroupGraph as GroupGraphData, GroupItem } from '@/api/client';
import { GraphLegend } from '@/features/group-graph/GraphLegend';
import { toGraphElements } from '@/features/group-graph/graphElements';
import { highlightLinksOnTap } from '@/features/group-graph/graphHighlight';
import { graphLayout, graphPaddingPx } from '@/features/group-graph/graphLayout';
import { graphStylesheet } from '@/features/group-graph/graphStyle';
import { useGroupGraph } from '@/features/group-graph/useGroupGraph';
import classes from '@/features/group-graph/GroupGraph.module.css';

type Props = { groupId: string; scheme: GroupItem['scheme'] };
type RendererWithSyncRender = { renderer?: () => { render?: () => void } };

const redrawBeforePaint = (cy: Core) =>
  (cy as unknown as RendererWithSyncRender).renderer?.().render?.();

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
  const theme = useMantineTheme();

  useEffect(() => {
    const cy = cytoscape({
      container: container.current,
      elements: toGraphElements(scheme, graph),
      style: graphStylesheet(theme),
      autounselectify: true,
      boxSelectionEnabled: false,
    });
    cy.layout(graphLayout(scheme, cy)).run();
    highlightLinksOnTap(cy);

    const refitOnResize = new ResizeObserver(() => {
      cy.resize();
      cy.fit(undefined, graphPaddingPx);
      redrawBeforePaint(cy);
    });
    refitOnResize.observe(container.current!);

    return () => {
      refitOnResize.disconnect();
      cy.destroy();
    };
  }, [scheme, graph, theme]);

  return (
    <div className={classes.canvasArea}>
      <div ref={container} className={classes.canvas} role="img" aria-label="Граф группы" />
    </div>
  );
}
