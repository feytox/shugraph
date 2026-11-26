import { Alert, Button, Code, Group, Loader, Paper, Stack, Text } from '@mantine/core';
import { IconPlayerPlay } from '@tabler/icons-react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { listScenarios, runScenario } from './scenarioRequests';
import classes from './ScenarioList.module.css';

function ScenarioRow({ id }: { id: string }) {
  const launch = useMutation({ mutationFn: runScenario, networkMode: 'always' });

  return (
    <li className={classes.row}>
      <Group justify="space-between" align="center">
        <Code className={classes.scenarioId}>{id}</Code>
        <Button
          aria-label={`Запустить ${id}`}
          leftSection={<IconPlayerPlay size={16} aria-hidden />}
          loading={launch.isPending}
          onClick={() => launch.mutate(id)}
        >
          Запустить
        </Button>
      </Group>
      {launch.isSuccess && (
        <Text role="status" c="green.8" size="sm" mt="xs">
          Запуск принят. События отправляются в фоновом режиме.
        </Text>
      )}
      {launch.isError && (
        <Text role="alert" c="red.8" size="sm" mt="xs">
          {launch.error.message}
        </Text>
      )}
    </li>
  );
}

export function ScenarioList() {
  const scenarios = useQuery({
    queryKey: ['scenarios'],
    queryFn: ({ signal }) => listScenarios(signal),
    networkMode: 'always',
    refetchOnWindowFocus: false,
    retry: false,
  });

  if (scenarios.isPending) {
    return (
      <Group role="status" gap="xs">
        <Loader size="xs" aria-hidden />
        <Text size="sm">Загружаем сценарии…</Text>
      </Group>
    );
  }

  if (scenarios.isError) {
    return (
      <Alert color="red" aria-busy={scenarios.isFetching}>
        <Stack gap="sm" align="flex-start">
          <Text size="sm">Не удалось загрузить сценарии.</Text>
          <Button
            variant="light"
            loading={scenarios.isFetching}
            onClick={() => scenarios.refetch()}
          >
            Повторить
          </Button>
        </Stack>
      </Alert>
    );
  }

  if (scenarios.data.length === 0) {
    return <Text c="dimmed">Нет доступных сценариев.</Text>;
  }

  return (
    <Paper withBorder radius="md">
      <ul aria-label="Сценарии" className={classes.list}>
        {scenarios.data.map((id) => (
          <ScenarioRow key={id} id={id} />
        ))}
      </ul>
    </Paper>
  );
}
