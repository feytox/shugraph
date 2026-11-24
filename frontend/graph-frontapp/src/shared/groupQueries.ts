export const groupsListKey = ['groups'];
export const groupsPollIntervalMs = 3000;

export const groupDetailKey = (groupId: string) => [...groupsListKey, groupId, 'detail'];
