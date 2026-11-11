import { MantineProvider } from '@mantine/core';
import { GroupsPanel } from '@/features/groups-list/GroupsPanel';

export function App() {
  return (
    <MantineProvider>
      <GroupsPanel />
    </MantineProvider>
  );
}
