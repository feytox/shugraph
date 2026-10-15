import { useState } from 'react';
import {
  MantineProvider,
  Container,
  Title,
  Paper,
  TextInput,
  Button,
  Group,
  Stack,
  Text,
  Alert,
  Loader,
  Badge,
  Card,
  Box,
} from '@mantine/core';
import { IconAlertCircle, IconCheck, IconRefresh, IconPlus } from '@tabler/icons-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import { apiClient, type Note } from '@/api/client';

dayjs.extend(utc);

export function NotesApp() {
  const [inputText, setInputText] = useState('');
  const [formError, setFormError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const {
    data: notes,
    isLoading,
    isError,
    error,
    refetch,
    isFetching,
  } = useQuery<Note[]>({
    queryKey: ['notes'],
    queryFn: async () => {
      const response = await apiClient.GET('/api/v1/notes');
      if (!response.response.ok || response.error !== undefined || !response.data) {
        throw new Error('Не удалось загрузить список записей с сервера');
      }
      return response.data.notes ?? [];
    },
  });

  const createMutation = useMutation({
    mutationFn: async (text: string) => {
      const { data, error, response } = await apiClient.POST('/api/v1/notes', {
        body: { text },
      });
      if (!response.ok) {
        const detail = typeof error === 'object' ? error?.detail : undefined;
        throw new Error(detail || `Ошибка при создании записи (HTTP ${response.status})`);
      }
      return data;
    },
    onSuccess: () => {
      setInputText('');
      setFormError(null);
      queryClient.invalidateQueries({ queryKey: ['notes'] });
    },
    onError: (err: Error) => {
      setFormError(err.message);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputText.trim()) {
      setFormError('Введите текст перед созданием');
      return;
    }
    setFormError(null);
    createMutation.mutate(inputText.trim());
  };

  return (
    <Container size="sm" py="xl">
      <Stack gap="lg">
        <Box>
          <Title order={2}>Тестовая страница (GSD-5 / B1)</Title>
          <Text c="dimmed" size="sm">
            Проверка взаимодействия с API записей (GET/POST /api/v1/notes)
          </Text>
        </Box>

        <Paper withBorder p="md" radius="md">
          <form onSubmit={handleSubmit}>
            <Stack gap="sm">
              <TextInput
                label="Текст записи"
                placeholder="Например: подозрительный перевод со счета 123"
                value={inputText}
                onChange={(e) => setInputText(e.currentTarget.value)}
                error={formError}
                disabled={createMutation.isPending}
                aria-label="Текст записи"
              />
              <Group justify="flex-end">
                <Button
                  type="submit"
                  leftSection={<IconPlus size={16} />}
                  loading={createMutation.isPending}
                >
                  Создать
                </Button>
              </Group>
            </Stack>
          </form>
        </Paper>

        <Paper withBorder p="md" radius="md">
          <Stack gap="md">
            <Group justify="space-between">
              <Group gap="xs">
                <Title order={4}>Список записей</Title>
                {notes && <Badge variant="light">{notes.length}</Badge>}
              </Group>
              <Button
                variant="light"
                leftSection={<IconRefresh size={16} />}
                onClick={() => refetch()}
                loading={isFetching}
              >
                Обновить
              </Button>
            </Group>

            {isError && (
              <Alert
                icon={<IconAlertCircle size={16} />}
                title="Ошибка подключения"
                color="red"
                variant="light"
              >
                {error?.message ||
                  'Сервис бэкенда недоступен. Проверьте запуск бэкенда или используйте моки.'}
              </Alert>
            )}

            {isLoading && (
              <Group justify="center" p="md">
                <Loader size="sm" />
                <Text size="sm" c="dimmed">
                  Загрузка данных...
                </Text>
              </Group>
            )}

            {!isLoading && !isError && notes && notes.length === 0 && (
              <Text c="dimmed" size="sm" ta="center" py="md">
                Записей пока нет. Создайте первую запись выше.
              </Text>
            )}

            {!isLoading && !isError && notes && notes.length > 0 && (
              <Stack gap="xs">
                {notes.map((note) => (
                  <Card key={note.id} withBorder padding="sm" radius="sm">
                    <Group justify="space-between" align="flex-start">
                      <Stack gap={4}>
                        <Text size="sm" fw={500}>
                          {note.text}
                        </Text>
                        <Text size="xs" c="dimmed">
                          UTC: {dayjs.utc(note.created_at).format('YYYY-MM-DD HH:mm:ss')} (id:{' '}
                          {note.id})
                        </Text>
                      </Stack>
                      <IconCheck size={16} color="green" />
                    </Group>
                  </Card>
                ))}
              </Stack>
            )}
          </Stack>
        </Paper>
      </Stack>
    </Container>
  );
}

export function App() {
  return (
    <MantineProvider>
      <NotesApp />
    </MantineProvider>
  );
}
