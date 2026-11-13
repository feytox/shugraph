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
type Props = {
  side: Side;
  title: string;
  storageKey: string;
  footer?: ReactNode;
  children: ReactNode;
};

const icons = {
  left: { collapse: IconLayoutSidebarLeftCollapse, expand: IconLayoutSidebarLeftExpand },
  right: { collapse: IconLayoutSidebarRightCollapse, expand: IconLayoutSidebarRightExpand },
};

export function SidePanel({ side, title, storageKey, footer, children }: Props) {
  const [collapsed, setCollapsed] = useLocalStorage({
    key: `${storageKey}-collapsed`,
    defaultValue: false,
  });
  const { collapse: CollapseIcon, expand: ExpandIcon } = icons[side];
  const ToggleIcon = collapsed ? ExpandIcon : CollapseIcon;

  return (
    <aside
      aria-label={title}
      className={classes.panel}
      data-side={side}
      data-collapsed={collapsed || undefined}
    >
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
          size="md"
          aria-label={`${collapsed ? 'Показать' : 'Скрыть'}: ${title}`}
          onClick={() => setCollapsed(!collapsed)}
        >
          <ToggleIcon size={18} />
        </ActionIcon>
        <Title
          order={2}
          size="h4"
          lineClamp={1}
          className={classes.fadesWhenCollapsed}
          aria-hidden={collapsed}
        >
          {title}
        </Title>
      </Group>
      <ScrollArea
        type="never"
        className={`${classes.body} ${classes.fadesWhenCollapsed}`}
        inert={collapsed}
      >
        {children}
      </ScrollArea>
      <div className={`${classes.footer} ${classes.fadesWhenCollapsed}`} inert={collapsed}>
        {footer}
      </div>
    </aside>
  );
}
