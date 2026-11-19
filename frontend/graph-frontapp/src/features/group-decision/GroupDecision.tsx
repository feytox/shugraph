import { Button, Divider, Group, Stack, Text, Textarea } from '@mantine/core';
import { useState } from 'react';
import type { GroupSolutionInput } from '@/api/client';
import { useGroupDecision } from '@/features/group-decision/useGroupDecision';
import { ErrorNotice } from '@/shared/ErrorNotice';

export function GroupDecision({ groupId }: { groupId: string }) {
  const [comment, setComment] = useState('');
  const { save, pending, error } = useGroupDecision(groupId);
  const decisionComment = comment.trim();

  const submit = (status: GroupSolutionInput['status']) => {
    if (!decisionComment) return;
    save({ status, decision_comment: decisionComment });
  };

  return (
    <Stack gap="xs">
      <Divider />
      <Textarea
        label="Комментарий"
        required
        minRows={3}
        disabled={pending}
        value={comment}
        onChange={(event) => setComment(event.currentTarget.value)}
      />
      <Group grow gap="xs">
        <Button
          size="compact-sm"
          disabled={!decisionComment || pending}
          onClick={() => submit('approved')}
        >
          Подтвердить
        </Button>
        <Button
          size="compact-sm"
          variant="default"
          disabled={!decisionComment || pending}
          onClick={() => submit('declined')}
        >
          Отклонить
        </Button>
      </Group>
      {pending && (
        <Text size="sm" role="status">
          Сохранение решения…
        </Text>
      )}
      {error && <ErrorNotice title={error} />}
    </Stack>
  );
}
