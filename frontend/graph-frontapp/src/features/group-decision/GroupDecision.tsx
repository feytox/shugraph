import { Button, Divider, Group, Stack, Text, Textarea } from '@mantine/core';
import { useState } from 'react';
import type { GroupSolutionInput } from '@/api/client';
import { useGroupDecision } from '@/features/group-decision/useGroupDecision';
import { ErrorNotice } from '@/shared/ErrorNotice';

export function GroupDecision({
  groupId,
  isOpen,
  evidenceReady,
}: {
  groupId: string;
  isOpen: boolean;
  evidenceReady: boolean;
}) {
  const [draft, setDraft] = useState<string>();
  const { save, pending, error, submittedComment } = useGroupDecision(groupId);
  const comment = draft ?? submittedComment ?? '';
  const decisionComment = comment.trim();
  const canSubmit = evidenceReady && Boolean(decisionComment) && !pending;
  const errorNotice = error && <ErrorNotice title={error} wrapText />;

  if (!isOpen) return errorNotice;

  const submit = (status: GroupSolutionInput['status']) => {
    if (!canSubmit) return;
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
        onChange={(event) => setDraft(event.currentTarget.value)}
      />
      <Group grow gap="xs">
        <Button size="compact-sm" disabled={!canSubmit} onClick={() => submit('approved')}>
          Подтвердить
        </Button>
        <Button
          size="compact-sm"
          variant="default"
          disabled={!canSubmit}
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
      {errorNotice}
    </Stack>
  );
}
