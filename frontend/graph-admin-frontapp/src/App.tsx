import { Box, Container, Stack, Text, Title } from '@mantine/core';
import { ScenarioList } from '@/features/scenarios/ScenarioList';
import classes from './App.module.css';

export function App() {
  return (
    <Box className={classes.page}>
      <Box component="header" className={classes.header}>
        <Container size="sm" py="md">
          <Text fw={650}>ShuGraph Admin</Text>
        </Container>
      </Box>
      <Container component="main" size="sm" py="xl">
        <Stack gap="lg">
          <Title order={1}>Запуск сценариев</Title>
          <ScenarioList />
        </Stack>
      </Container>
    </Box>
  );
}
