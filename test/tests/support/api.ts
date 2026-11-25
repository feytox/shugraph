import { expect, type APIRequestContext, type APIResponse } from '@playwright/test';
import type { ExpectedGroupDetails, Scheme, ScenarioEvent } from './scenarios';

export type GroupItem = {
  id: string;
  scheme: Scheme;
  status: 'new' | 'approved' | 'declined';
  risk_level: 'low' | 'medium' | 'high';
  risk_score: number;
  risk_parts: Partial<Record<Scheme, Record<string, number>>>;
  detected_at: string;
  first_event_at: string;
  last_event_at: string;
  features: {
    accounts_count: number;
    span_minutes: number;
    total_amount: number;
  };
};

export type GroupDetails = GroupItem & {
  is_frozen: boolean;
  decision?: { comment: string; decided_at: string };
  members: { id: string; client_id: string }[];
  logins: {
    event_id: string;
    occurred_at: string;
    account_id: string;
    client_id: string;
    device_id: string;
    address_id: string;
  }[];
  transfers: {
    event_id: string;
    occurred_at: string;
    amount: number;
    from_account_id: string;
    from_client_id: string;
    to_account_id: string;
    to_client_id: string;
  }[];
};

export type GraphNode = { id: string; type: string; entity_id: string };
export type GraphEdge = { id: string; source: string; target: string; event_id: string };

export type GroupGraph = {
  group_id: string;
  nodes: GraphNode[];
  ownerships: GraphEdge[];
  logins: GraphEdge[];
  transfers: (GraphEdge & { amount: number; occurred_at: string })[];
};

const groupsPath = '/api/v1/groups';

export function sleep(milliseconds: number) {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

export async function runScenario(
  request: APIRequestContext,
  name: string,
  delaySeconds = 0
): Promise<void> {
  const response = await request.post(`/admin-api/v1/scenarios/${name}/run`, {
    params: { delay_s: String(delaySeconds) },
  });
  expect(response.status(), `run scenario ${name}: ${await response.text()}`).toBe(202);
}

export async function listGroups(request: APIRequestContext): Promise<GroupItem[]> {
  const items: GroupItem[] = [];
  let cursor: string | undefined;

  do {
    const response = await request.get(groupsPath, {
      params: { limit: '100', ...(cursor ? { cursor } : {}) },
    });
    expect(response.status(), `list groups: ${await response.text()}`).toBe(200);
    const page = (await response.json()) as { items: GroupItem[]; next_cursor: string | null };
    items.push(...page.items);
    cursor = page.next_cursor ?? undefined;
  } while (cursor);

  return items;
}

export function matchesGroup(group: GroupItem, expected: ExpectedGroupDetails): boolean {
  return (
    group.scheme === expected.scheme &&
    Date.parse(group.first_event_at) === Date.parse(expected.firstEventAt) &&
    Date.parse(group.last_event_at) === Date.parse(expected.lastEventAt) &&
    group.features.accounts_count === expected.accountsCount &&
    group.features.total_amount === expected.totalAmount
  );
}

export async function waitForGroups(
  request: APIRequestContext,
  expected: ExpectedGroupDetails[],
  timeoutMs = 30_000
): Promise<GroupItem[]> {
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    const groups = await listGroups(request);
    const matched = expected.map((item) => groups.find((group) => matchesGroup(group, item)));
    if (matched.every(Boolean)) return matched as GroupItem[];
    await sleep(500);
  }

  const groups = await listGroups(request);
  throw new Error(
    `groups for ${expected.map((item) => item.scheme).join(', ')} not found; ` +
      `visible groups: ${groups.map((group) => `${group.scheme}@${group.last_event_at}`).join(', ')}`
  );
}

export async function ensureScenarioGroups(
  request: APIRequestContext,
  scenario: string,
  expected: ExpectedGroupDetails[]
): Promise<GroupItem[]> {
  await runScenario(request, scenario);
  return waitForGroups(request, expected);
}

export async function getGroup(
  request: APIRequestContext,
  groupId: string
): Promise<GroupDetails> {
  const response = await request.get(`${groupsPath}/${groupId}`);
  expect(response.status(), `get group: ${await response.text()}`).toBe(200);
  return (await response.json()) as GroupDetails;
}

export async function getGraph(
  request: APIRequestContext,
  groupId: string
): Promise<GroupGraph> {
  const response = await request.get(`${groupsPath}/${groupId}/graph`);
  expect(response.status(), `get group graph: ${await response.text()}`).toBe(200);
  return (await response.json()) as GroupGraph;
}

export async function patchSolution(
  request: APIRequestContext,
  groupId: string,
  status: 'approved' | 'declined',
  decisionComment: string
): Promise<APIResponse> {
  return request.patch(`${groupsPath}/${groupId}/solution`, {
    data: { status, decision_comment: decisionComment },
  });
}

export async function putScenarioEvent(
  request: APIRequestContext,
  event: ScenarioEvent,
  override: Partial<ScenarioEvent> = {}
): Promise<APIResponse> {
  const merged = { ...event, ...override };
  const { event_id: eventId, type, ...payload } = merged;
  const path =
    type === 'login'
      ? `/api/v1/events/logins/${encodeURIComponent(eventId)}`
      : `/api/v1/events/transfers/${encodeURIComponent(eventId)}`;

  return request.put(path, { data: payload });
}
