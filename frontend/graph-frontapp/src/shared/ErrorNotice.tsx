import { ActionIcon, Group, Text } from '@mantine/core';
import { IconRefresh } from '@tabler/icons-react';
import classes from '@/shared/ErrorNotice.module.css';

type Props = {
  title: string;
  details?: string;
  onRetry?: () => void;
  retrying?: boolean;
};

export function ErrorNotice({ title, details, onRetry, retrying = false }: Props) {
  return (
    <Group role="alert" gap="xs" wrap="nowrap" className={classes.notice}>
      <Text size="sm" fw={600} className={classes.title}>
        {title}
      </Text>
      {details && (
        <Text size="sm" truncate>
          {details}
        </Text>
      )}
      {onRetry && (
        <ActionIcon
          variant="transparent"
          color="white"
          size="sm"
          aria-label="Повторить"
          loading={retrying}
          onClick={onRetry}
        >
          <IconRefresh size={14} />
        </ActionIcon>
      )}
    </Group>
  );
}
