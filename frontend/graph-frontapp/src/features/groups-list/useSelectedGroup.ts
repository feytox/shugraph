import { useEffect } from 'react';
import { useSearchParams } from 'react-router';
import type { GroupItem } from '@/api/client';
import { useGroups } from '@/features/groups-list/useGroups';

const selectedGroupParam = 'group';

export const searchWithGroup = (params: URLSearchParams, groupId: string) => {
  const next = new URLSearchParams(params);
  next.set(selectedGroupParam, groupId);
  return `?${next}`;
};

export function useSelectedGroupId() {
  const [params] = useSearchParams();
  return params.get(selectedGroupParam);
}

export function useDefaultGroupSelection() {
  const [, setParams] = useSearchParams();
  const selectedId = useSelectedGroupId();
  const firstId = useGroups().data?.items[0]?.id;

  useEffect(() => {
    if (!selectedId && firstId) {
      setParams((current) => new URLSearchParams(searchWithGroup(current, firstId)), {
        replace: true,
      });
    }
  }, [selectedId, firstId, setParams]);
}

export function useSelectedGroup(): { group?: GroupItem; notInList: boolean } {
  const selectedId = useSelectedGroupId();
  const { data } = useGroups();
  const group = data?.items.find((item) => item.id === selectedId);
  return { group, notInList: Boolean(selectedId && data && !group) };
}
