import { Group, Paper, Text, useMantineTheme } from '@mantine/core';
import { nodeColor, nodeKinds } from '@/features/group-graph/graphStyle';
import classes from '@/features/group-graph/GroupGraph.module.css';

export function GraphLegend() {
  const theme = useMantineTheme();
  return (
    <Paper withBorder px="sm" py={4} className={classes.legend}>
      <Group gap="md">
        {(Object.keys(nodeKinds) as (keyof typeof nodeKinds)[]).map((type) => {
          const { label, Icon } = nodeKinds[type];
          return (
            <Group key={type} gap={4}>
              <Icon size={12} color={nodeColor(theme, type)} aria-hidden />
              <Text size="xs" c="dimmed">
                {label}
              </Text>
            </Group>
          );
        })}
        <Group gap={4}>
          <span className={classes.transferLine} />
          <Text size="xs" c="dimmed">
            перевод
          </Text>
        </Group>
        <Group gap={4}>
          <span className={classes.loginLine} />
          <Text size="xs" c="dimmed">
            вход
          </Text>
        </Group>
        <Text size="xs" c="dimmed">
          время в UTC
        </Text>
      </Group>
    </Paper>
  );
}
