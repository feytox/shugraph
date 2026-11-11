import type { ReactNode } from 'react';
import { Box, Group, Text } from '@mantine/core';
import { nodeKinds, shownNodeTypes } from '@/features/group-graph/nodeKinds';
import classes from '@/features/group-graph/GroupGraph.module.css';

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

export function GraphLegend() {
  return (
    <Group justify="space-between" wrap="nowrap" px="md" py="xs" className={classes.legend}>
      <Group gap="lg" wrap="nowrap">
        {shownNodeTypes.map((type) => {
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
        <LegendItem
          label="перевод"
          marker={<span className={classes.transferLine} aria-hidden />}
        />
        <LegendItem label="вход" marker={<span className={classes.loginLine} aria-hidden />} />
      </Group>
      <Text size="xs" c="dimmed">
        время в UTC
      </Text>
    </Group>
  );
}
