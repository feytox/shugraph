import { useRef, useState, type KeyboardEvent, type PointerEvent, type ReactNode } from 'react';
import { ActionIcon, Box, Group, ScrollArea, Title } from '@mantine/core';
import { useLocalStorage } from '@mantine/hooks';
import {
  IconLayoutSidebarLeftCollapse,
  IconLayoutSidebarLeftExpand,
  IconLayoutSidebarRightCollapse,
  IconLayoutSidebarRightExpand,
} from '@tabler/icons-react';
import { clampPanelWidth, panelWidthPx } from '@/shared/panelWidth';
import classes from '@/shared/SidePanel.module.css';

type Side = 'left' | 'right';
type Props = { side: Side; title: string; storageKey: string; children: ReactNode };

const icons = {
  left: { collapse: IconLayoutSidebarLeftCollapse, expand: IconLayoutSidebarLeftExpand },
  right: { collapse: IconLayoutSidebarRightCollapse, expand: IconLayoutSidebarRightExpand },
};

const growsWhenDraggedRight: Record<Side, 1 | -1> = { left: 1, right: -1 };

export function SidePanel({ side, title, storageKey, children }: Props) {
  const [width, setWidth] = useLocalStorage({
    key: `${storageKey}-width`,
    defaultValue: panelWidthPx.initial,
  });
  const [collapsed, setCollapsed] = useLocalStorage({
    key: `${storageKey}-collapsed`,
    defaultValue: false,
  });
  const dragStart = useRef<{ x: number; width: number } | null>(null);
  const [dragging, setDragging] = useState(false);
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

  const startDrag = (event: PointerEvent<HTMLDivElement>) => {
    event.preventDefault();
    event.currentTarget.setPointerCapture?.(event.pointerId);
    dragStart.current = { x: event.clientX, width };
    setDragging(true);
  };
  const drag = (event: PointerEvent<HTMLDivElement>) => {
    if (!dragStart.current) return;
    const shift = (event.clientX - dragStart.current.x) * growsWhenDraggedRight[side];
    setWidth(clampPanelWidth(dragStart.current.width + shift));
  };
  const stopDrag = () => {
    dragStart.current = null;
    setDragging(false);
  };
  const resizeWithKeys = (event: KeyboardEvent<HTMLDivElement>) => {
    const direction = { ArrowRight: 1, ArrowLeft: -1 }[event.key];
    if (!direction) return;
    event.preventDefault();
    const step = direction * growsWhenDraggedRight[side] * panelWidthPx.keyboardStep;
    setWidth(clampPanelWidth(width + step));
  };

  return (
    <Box component="aside" aria-label={title} className={classes.panel} data-side={side} w={width}>
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
        <Title order={2} size="h4">
          {title}
        </Title>
      </Group>
      <ScrollArea type="scroll" className={classes.body}>
        {children}
      </ScrollArea>
      <div
        role="separator"
        aria-orientation="vertical"
        aria-label={`Ширина: ${title}`}
        aria-valuenow={width}
        aria-valuemin={panelWidthPx.min}
        aria-valuemax={panelWidthPx.max}
        tabIndex={0}
        className={classes.resizer}
        data-dragging={dragging || undefined}
        onPointerDown={startDrag}
        onPointerMove={drag}
        onPointerUp={stopDrag}
        onPointerCancel={stopDrag}
        onKeyDown={resizeWithKeys}
      />
    </Box>
  );
}
