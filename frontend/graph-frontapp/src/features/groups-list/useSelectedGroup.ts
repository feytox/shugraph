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

export function useSelectedGroup(): GroupItem | undefined {
  const [params, setParams] = useSearchParams();
  const { data } = useGroups();
  const items = data?.items ?? [];
  const selectedId = params.get(selectedGroupParam);
  const firstId = items[0]?.id;

  useEffect(() => {
    if (!selectedId && firstId) {
      setParams((current) => new URLSearchParams(searchWithGroup(current, firstId)), {
        replace: true,
      });
    }
  }, [selectedId, firstId, setParams]);

  return items.find((group) => group.id === selectedId);
}
