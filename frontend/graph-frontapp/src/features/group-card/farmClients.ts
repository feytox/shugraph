import type { Group } from '@/api/client';

export const farmClientCount = (group: Group) =>
  new Set(group.logins.map(({ client_id }) => client_id)).size;
