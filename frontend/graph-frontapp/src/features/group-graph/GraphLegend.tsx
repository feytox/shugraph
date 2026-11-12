import type { ReactNode } from 'react';
import { Box, Group, Text } from '@mantine/core';
import type { ElementsDefinition } from 'cytoscape';
import { nodeKinds, shownNodeTypes } from '@/features/group-graph/nodeKinds';
import classes from '@/features/group-graph/GroupGraph.module.css';

const edgeKinds = [
  { type: 'transfer', label: 'перевод', markerClass: classes.transferLine },
  { type: 'login', label: 'вход', markerClass: classes.loginLine },
];

function LegendItem({ marker, label }: { marker: ReactNode; label: string }) {
  return (
    <Group gap="xs" wrap="nowrap">
      {marker}
      <Text size="xs" c="dimmed">
        {label}
      </Text>
    </Group>
  );
}

export function GraphLegend({
  elements,
  notice,
}: {
  elements?: ElementsDefinition;
  notice?: ReactNode;
}) {
  const hasNodeType = (type: string) => !!elements?.nodes.some((node) => node.data.type === type);
  const hasEdgeType = (type: string) => !!elements?.edges.some((edge) => edge.data.type === type);

  return (
    <Group justify="space-between" wrap="nowrap" px="md" py="xs" className={classes.legend}>
      <Group gap="lg" wrap="nowrap">
        {shownNodeTypes.filter(hasNodeType).map((type) => {
          const { label, color, legendMarker } = nodeKinds[type];
          return (
            <LegendItem
              key={type}
              label={label}
              marker={
                <Box
                  component="span"
                  bg={`${color}.6`}
                  className={classes[legendMarker]}
                  aria-hidden
                />
              }
            />
          );
        })}
        {edgeKinds
          .filter(({ type }) => hasEdgeType(type))
          .map(({ type, label, markerClass }) => (
            <LegendItem
              key={type}
              label={label}
              marker={<span className={markerClass} aria-hidden />}
            />
          ))}
      </Group>
      {notice}
      <Text size="xs" c="dimmed">
        время в UTC
      </Text>
    </Group>
  );
}
