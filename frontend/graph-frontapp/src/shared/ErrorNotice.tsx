import { ActionIcon, Group, Text } from '@mantine/core';
import { IconRefresh } from '@tabler/icons-react';
import classes from '@/shared/ErrorNotice.module.css';

type Props = {
  title: string;
  onRetry?: () => void;
  retrying?: boolean;
};

export function ErrorNotice({ title, onRetry, retrying = false }: Props) {
  return (
    <Group role="alert" gap="xs" wrap="nowrap" className={classes.notice}>
      <Text size="sm" fw={600} truncate className={classes.title}>
        {title}
      </Text>
      {onRetry && (
        <ActionIcon
          variant="subtle"
          color="red"
          size="sm"
          aria-label={`Повторить: ${title}`}
          loading={retrying}
          onClick={onRetry}
        >
          <IconRefresh size={14} />
        </ActionIcon>
      )}
    </Group>
  );
}
