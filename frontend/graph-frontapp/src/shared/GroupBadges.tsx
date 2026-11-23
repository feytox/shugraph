import { Badge } from '@mantine/core';
import type { GroupItem } from '@/api/client';
import { riskColors, riskLabels, statusColors, statusLabels } from '@/shared/groupLabels';
import classes from '@/shared/GroupBadges.module.css';

export function RiskBadge({ level, score }: { level: GroupItem['risk_level']; score: number }) {
  return (
    <Badge
      color={riskColors[level]}
      variant="filled"
      size="lg"
      fw={800}
      className={classes.keepWidth}
      aria-label={`Риск ${riskLabels[level].toLowerCase()}: ${score}`}
    >
      {score}
    </Badge>
  );
}

export function StatusBadge({ status }: { status: GroupItem['status'] }) {
  return (
    <Badge
      color={statusColors[status]}
      variant="light"
      size="sm"
      tt="none"
      className={classes.keepWidth}
    >
      {statusLabels[status]}
    </Badge>
  );
}
