import { expect, test } from '@playwright/test';
import {
  ensureScenarioGroups,
  getGraph,
  getGroup,
  listGroups,
  matchesGroup,
  patchSolution,
  putScenarioEvent,
  runScenario,
  sleep,
} from '../support/api';
import { expectedGroup, loadScenario } from '../support/scenarios';

test('device_farm: список, карточка и граф отдают данные сценария', async ({ request }) => {
  const expected = expectedGroup('device_farm');
  const [group] = await ensureScenarioGroups(request, 'device_farm', [expected]);

  expect(group.scheme).toBe('farm');
  expect(group.risk_score).toBe(75);
  expect(group.risk_level).toBe('high');
  expect(group.risk_parts.farm).toEqual({
    scheme: 50,
    speed: 15,
    clients: 0,
    collector: 10,
  });
  expect(group.features).toEqual({
    accounts_count: 4,
    span_minutes: 28,
    total_amount: 16_100_000,
  });
  expect(group.first_event_at).toBe('2026-11-05T10:01:00Z');
  expect(group.last_event_at).toBe('2026-11-05T10:29:00Z');

  const details = await getGroup(request, group.id);

  expect(details.members.map((member) => member.id).sort()).toEqual(
    [...expected.member_account_ids].sort()
  );
  expect(details.logins).toHaveLength(3);
  expect(details.transfers).toHaveLength(3);
  expect(details.transfers.map((transfer) => transfer.amount).sort((a, b) => a - b)).toEqual([
    4_200_000, 5_200_000, 6_700_000,
  ]);
  expect(typeof details.is_frozen).toBe('boolean');
  if (group.status !== 'new') expect(details.is_frozen).toBe(true);

  const graph = await getGraph(request, group.id);

  expect(graph.group_id).toBe(group.id);
  expect(graph.nodes).toHaveLength(12);
  expect(graph.logins).toHaveLength(6);
  expect(graph.transfers).toHaveLength(3);
  expect(graph.nodes.map((node) => node.id).sort()).toEqual(
    [
      'client:cli-ravil',
      'client:cli-sasha',
      'client:cli-dima',
      'client:cli-organizer',
      'account:acc-ravil',
      'account:acc-sasha',
      'account:acc-dima',
      'account:acc-collector',
      'device:dev-common',
      'network_address:198.51.100.21',
      'network_address:198.51.100.22',
      'network_address:198.51.100.23',
    ].sort()
  );
});

for (const scenario of ['farm_and_address', 'farm_and_chain'] as const) {
  test(`${scenario}: ферма появляется с ожидаемыми числами`, async ({ request }) => {
    const expected = expectedGroup(scenario);
    const [group] = await ensureScenarioGroups(request, scenario, [expected]);

    expect(group.scheme).toBe('farm');
    expect(group.risk_score).toBe(expected.risk_score);
    expect(group.risk_level).toBe(expected.risk_level);
    expect(group.risk_parts.farm).toEqual(expected.riskParts);
    expect(group.features.accounts_count).toBe(expected.accountsCount);
    expect(group.features.span_minutes).toBe(expected.spanMinutes);
    expect(group.features.total_amount).toBe(expected.totalAmount);
  });
}

test('решение: сценарий → подтверждение → повторное решение 409', async ({ request }) => {
  const scenario = 'farm_and_chain';
  const expectedFarm = expectedGroup(scenario);
  const [group] = await ensureScenarioGroups(request, scenario, [expectedFarm]);

  const current = await getGroup(request, group.id);
  const comment = 'API E2E: ферма подтверждена';

  if (current.status === 'new') {
    const response = await patchSolution(request, group.id, 'approved', comment);
    expect(response.status()).toBe(200);
    const body = (await response.json()) as { status: string; decision: { comment: string } };
    expect(body.status).toBe('approved');
    expect(body.decision.comment).toBe(comment);
  }

  const resolved = await getGroup(request, group.id);
  expect(resolved.status).toBe('approved');
  expect(resolved.is_frozen).toBe(true);
  expect(resolved.decision?.comment).toBeTruthy();

  const repeated = await patchSolution(request, group.id, 'declined', 'повторное решение');
  expect(repeated.status()).toBe(409);
});

test('повторный запуск сценария не создаёт дубль и не меняет решение', async ({ request }) => {
  const scenario = 'farm_and_chain';
  const expectedFarm = expectedGroup(scenario);
  await ensureScenarioGroups(request, scenario, [expectedFarm]);

  const before = (await listGroups(request)).filter((group) =>
    matchesGroup(group, expectedFarm)
  );
  expect(before).toHaveLength(1);
  const statusBefore = before[0].status;

  await runScenario(request, scenario);
  await sleep(8_000);

  const after = (await listGroups(request)).filter((group) => matchesGroup(group, expectedFarm));
  expect(after).toHaveLength(1);
  expect(after[0].status).toBe(statusBefore);
});

test('события: повтор 204, изменённое тело 409', async ({ request }) => {
  await ensureScenarioGroups(request, 'device_farm', [expectedGroup('device_farm')]);
  const scenario = loadScenario('device_farm');

  const login = scenario.events.find((event) => event.event_id === 'evt-df-with-0001');
  const transfer = scenario.events.find((event) => event.event_id === 'evt-df-with-0004');
  expect(login).toBeTruthy();
  expect(transfer).toBeTruthy();

  expect((await putScenarioEvent(request, login!)).status()).toBe(204);
  expect((await putScenarioEvent(request, login!, { device_id: 'dev-other' })).status()).toBe(409);

  expect((await putScenarioEvent(request, transfer!)).status()).toBe(204);
  expect((await putScenarioEvent(request, transfer!, { amount: 1 })).status()).toBe(409);
});

test('without_group: новых групп нет', async ({ request }) => {
  const before = await listGroups(request);
  await runScenario(request, 'without_group');
  await sleep(8_000);
  const after = await listGroups(request);

  expect(after).toHaveLength(before.length);
});
