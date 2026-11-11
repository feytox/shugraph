import type { ReactNode } from 'react';
import { ActionIcon, Group, ScrollArea, Title } from '@mantine/core';
import { useLocalStorage } from '@mantine/hooks';
import {
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarLeftExpand,
  IconLayoutSidebarRightCollapse,
  IconLayoutSidebarRightExpand,
} from '@tabler/icons-react';
import classes from '@/shared/SidePanel.module.css';

type Side = 'left' | 'right';
type Props = { side: Side; title: string; storageKey: string; children: ReactNode };

const icons = {
  left: { collapse: IconLayoutSidebarLeftCollapse, expand: IconLayoutSidebarLeftExpand },
  right: { collapse: IconLayoutSidebarRightCollapse, expand: IconLayoutSidebarRightExpand },
};

export function SidePanel({ side, title, storageKey, children }: Props) {
  const [collapsed, setCollapsed] = useLocalStorage({
    key: `${storageKey}-collapsed`,
    defaultValue: false,
  });
  const { collapse: CollapseIcon, expand: ExpandIcon } = icons[side];

  if (collapsed) {
    return (
      <aside aria-label={title} className={classes.collapsed} data-side={side}>
        <ActionIcon
          variant="subtle"
          color="gray"
          aria-label={`Показать: ${title}`}
          onClick={() => setCollapsed(false)}
        >
          <ExpandIcon size={18} />
        </ActionIcon>
      </aside>
    );
  }

  return (
    <aside aria-label={title} className={classes.panel} data-side={side}>
      <Group
        gap="xs"
        wrap="nowrap"
        px="md"
        pt="md"
        pb="xs"
        className={classes.header}
        data-side={side}
      >
        <ActionIcon
          variant="subtle"
          color="gray"
          aria-label={`Скрыть: ${title}`}
          onClick={() => setCollapsed(true)}
        >
          <CollapseIcon size={18} />
        </ActionIcon>
        <Title order={2} size="h4" lineClamp={1}>
          {title}
        </Title>
      </Group>
      <ScrollArea type="never" className={classes.body}>
        {children}
      </ScrollArea>
    </aside>
  );
}
