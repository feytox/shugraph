import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import type { Group, GroupItem, GroupSolutionInput, GroupSolutionResponse } from '@/api/client';
import { mockGroup, mockGroups } from '@/mocks/groupGraphs';

dayjs.extend(utc);

const evolvingGroupNumber = 25;
const evolvingStartAmount = 12345;
const evolvingTransferAmount = 10000;
const generatedGroupId = (number: number) =>
  `00000000-0000-4000-8000-${String(number).padStart(12, '0')}`;

function demoFarm(item: GroupItem) {
  const { accounts_count, total_amount } = item.features;
  const accounts = Array.from({ length: accounts_count }, (_, index) => `acc-${item.id}-${index}`);
  const evolving = item.id === generatedGroupId(evolvingGroupNumber);
  const transferCount = evolving
    ? (total_amount - evolvingStartAmount) / evolvingTransferAmount + 1
    : 1;
  const lastEventAt = dayjs.utc(item.last_event_at);

  return mockGroup({
    item,
    events: {
      owners: Object.fromEntries(accounts.map((account) => [account, `cli-${account}`])),
      logins: accounts.map((account, index) => ({
        event: `login-${account}`,
        at: dayjs.utc(item.first_event_at).add(index, 'minute').toISOString(),
        account,
        device: `dev-${item.id}`,
        address: `198.51.100.${index + 1}`,
      })),
      transfers: accounts.slice(0, transferCount).map((account, index) => ({
        event: `transfer-${account}`,
        at: lastEventAt.subtract(transferCount - index - 1, 'second').toISOString(),
        from: account,
        to: accounts[accounts.length - 1],
        rub:
          (evolving ? (index === 0 ? evolvingStartAmount : evolvingTransferAmount) : total_amount) /
          100,
      })),
    },
  });
}

export function createGroupsDemo() {
  const decisions = new Map<string, { item: GroupItem; solution: GroupSolutionResponse }>();
  let startedAt: number | undefined;
  function elapsed() {
    startedAt ??= Date.now();
    return Date.now() - startedAt;
  }

  function getItems(): GroupItem[] {
    const step = Math.min(3, Math.floor(elapsed() / 10000));
    const generated: GroupItem[] = Array.from({ length: 25 + step }, (_, index) => {
      const number = 25 + step - index;
      const updated = number === evolvingGroupNumber;
      const scheme = (['farm', 'address', 'chain', 'ring'] as const)[(number - 1) % 4];
      const accountsCount = scheme === 'farm' ? Math.max(3, number) : number;
      const collector = updated && step > 0 ? 10 : 0;
      const farm = {
        scheme: 50,
        speed: 15,
        clients: Math.min(15, (accountsCount - 3) * 5),
        collector,
      };
      const riskScore =
        scheme === 'farm' ? farm.scheme + farm.speed + farm.clients + farm.collector : 60 + number;
      return {
        id: generatedGroupId(number),
        scheme,
        status: 'new',
        risk_level: riskScore >= 75 ? 'high' : 'medium',
        risk_score: riskScore,
        risk_parts: scheme === 'farm' ? { farm } : {},
        detected_at: `2026-10-07T10:00:${String(number).padStart(2, '0')}Z`,
        first_event_at: '2026-10-07T10:00:00Z',
        last_event_at: updated
          ? `2026-10-07T10:30:${String(45 + step).padStart(2, '0')}Z`
          : '2026-10-07T10:30:45Z',
        features: {
          accounts_count: accountsCount,
          total_amount: updated
            ? evolvingStartAmount + step * evolvingTransferAmount
            : number * 10000,
          span_minutes: 30,
        },
      };
    });
    return [...mockGroups.map((group) => group.item), ...generated].map(
      (item) => decisions.get(item.id)?.item ?? item
    );
  }

  function getItem(groupId: string) {
    return getItems().find(({ id }) => id === groupId);
  }

  function findMockGroup(item: GroupItem) {
    return (
      mockGroups.find((group) => group.item.id === item.id) ??
      (item.scheme === 'farm' ? demoFarm(item) : undefined)
    );
  }

  return {
    getItems,
    getItem,
    getGroup(groupId: string): Group | undefined {
      const item = getItem(groupId);
      if (!item) return undefined;
      const evidence = findMockGroup(item)?.evidence ?? {
        members: [],
        logins: [],
        transfers: [],
      };
      const solution = decisions.get(groupId)?.solution;
      return { ...item, ...evidence, is_frozen: Boolean(solution), decision: solution?.decision };
    },
    getGraph(groupId: string) {
      const item = getItem(groupId);
      return item && findMockGroup(item)?.graph;
    },
    saveSolution(item: GroupItem, input: GroupSolutionInput): GroupSolutionResponse {
      const solution = {
        status: input.status,
        decision: { comment: input.decision_comment, decided_at: dayjs.utc().toISOString() },
      };
      decisions.set(item.id, { item: { ...item, status: solution.status }, solution });
      return solution;
    },
  };
}
