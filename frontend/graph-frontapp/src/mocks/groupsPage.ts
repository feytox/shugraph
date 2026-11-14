import type { GroupItem, GroupListResponse } from '@/api/client';

export function groupsPage(items: GroupItem[], request: Request): GroupListResponse {
  const query = new URL(request.url).searchParams;
  const cursor = query.get('cursor');
  const start = cursor ? items.findIndex((item) => btoa(item.id) === cursor) + 1 : 0;
  const limit = Math.min(20, Math.max(1, Number(query.get('limit') ?? 20)));
  const page = items.slice(start, start + limit);
  return {
    items: page,
    next_cursor: start + limit < items.length ? btoa(page[page.length - 1].id) : null,
  };
}
