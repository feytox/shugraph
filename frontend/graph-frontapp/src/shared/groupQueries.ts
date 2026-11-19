export const groupsListKey = ['groups'];

export const groupDetailKey = (groupId: string) => [...groupsListKey, groupId, 'detail'];
