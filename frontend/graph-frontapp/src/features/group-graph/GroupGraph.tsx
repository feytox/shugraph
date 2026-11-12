import { useMemo, useRef, useState, type ReactNode } from 'react';
import { ActionIcon, Alert, Button, Center, Group, Loader, Stack, Text } from '@mantine/core';
import { IconRefresh } from '@tabler/icons-react';
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
  notice,
  children,
}: {
  elements?: ElementsDefinition;
  notice?: ReactNode;
  children: ReactNode;
}) {
  return (
    <div className={classes.root}>
      <div className={classes.canvasArea}>{children}</div>
      <GraphLegend elements={elements} notice={notice} />
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
    <GraphFrame
      elements={drawn}
      notice={drawn && graph.isError ? <RefreshError graph={graph} /> : undefined}
    >
      {drawn ? <GraphCanvas scheme={scheme} elements={drawn} /> : <GraphStatus graph={graph} />}
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
    <Group role="alert" gap="xs" wrap="nowrap" className={classes.refreshError}>
      <Text size="xs">Не удалось обновить граф</Text>
      <ActionIcon
        variant="transparent"
        color="white"
        size="sm"
        aria-label="Обновить граф"
        loading={graph.isFetching}
        onClick={() => graph.refetch()}
      >
        <IconRefresh size={14} />
      </ActionIcon>
    </Group>
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
