import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

export type Scheme = 'farm' | 'address' | 'chain' | 'ring';
export type RiskLevel = 'low' | 'medium' | 'high';

export type ScenarioEvent = {
  event_id: string;
  type: 'login' | 'transfer';
  occurred_at: string;
  amount?: number;
  account_id?: string;
  client_id?: string;
  device_id?: string;
  address_id?: string;
  from_account_id?: string;
  from_client_id?: string;
  to_account_id?: string;
  to_client_id?: string;
};

export type ExpectedGroup = {
  scheme: Scheme;
  member_account_ids: string[];
  event_ids: string[];
  risk_score: number;
  risk_level: RiskLevel;
  risk_parts: Record<string, Record<string, number>>;
};

export type Scenario = {
  name: string;
  events: ScenarioEvent[];
  expected: { groups: ExpectedGroup[] };
};

export type ExpectedGroupDetails = ExpectedGroup & {
  accountsCount: number;
  spanMinutes: number;
  totalAmount: number;
  firstEventAt: string;
  lastEventAt: string;
  riskParts: Record<string, number>;
};

const scenariosDir = fileURLToPath(
  new URL('../../../backend/graph-admin-api/scenarios/', import.meta.url)
);

export function loadScenario(name: string): Scenario {
  const file = path.join(scenariosDir, `${name}.json`);
  return JSON.parse(readFileSync(file, 'utf8')) as Scenario;
}

export function expectedGroups(scenarioName: string): ExpectedGroupDetails[] {
  const scenario = loadScenario(scenarioName);
  const events = new Map(scenario.events.map((event) => [event.event_id, event]));

  return scenario.expected.groups.map((group) => {
    const groupEvents = group.event_ids.map((id) => {
      const event = events.get(id);
      if (!event) throw new Error(`event ${id} is missing in scenario ${scenarioName}`);
      return event;
    });

    const firstEventAt = groupEvents
      .map((event) => event.occurred_at)
      .sort((left, right) => Date.parse(left) - Date.parse(right))[0];

    const lastEventAt = groupEvents
      .map((event) => event.occurred_at)
      .sort((left, right) => Date.parse(right) - Date.parse(left))[0];

    const totalAmount = groupEvents.reduce(
      (sum, event) => sum + (event.type === 'transfer' ? (event.amount ?? 0) : 0),
      0
    );

    return {
      ...group,
      accountsCount: group.member_account_ids.length,
      spanMinutes: Math.floor((Date.parse(lastEventAt) - Date.parse(firstEventAt)) / 60_000),
      totalAmount,
      firstEventAt,
      lastEventAt,
      riskParts: group.risk_parts[group.scheme],
    };
  });
}

export function expectedGroup(scenarioName: string, scheme: Scheme = 'farm'): ExpectedGroupDetails {
  const group = expectedGroups(scenarioName).find((item) => item.scheme === scheme);
  if (!group) throw new Error(`scenario ${scenarioName} has no ${scheme} group`);
  return group;
}
