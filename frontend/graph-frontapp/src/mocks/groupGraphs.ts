import type { GraphNode, GroupGraph, GroupItem } from '@/api/client';

type Login = { event: string; at: string; account: string; device: string; address: string };
type Transfer = { event: string; at: string; from: string; to: string; rub: number };
type Events = { owners: Record<string, string>; logins: Login[]; transfers: Transfer[] };

export type MockGroup = { item: GroupItem; graph: GroupGraph };

const node = (type: GraphNode['type'], entityId: string): GraphNode => ({
  id: `${type}:${entityId}`,
  type,
  entity_id: entityId,
});

const unique = (values: string[]) => [...new Set(values)];

function buildGraph(groupId: string, { owners, logins, transfers }: Events): GroupGraph {
  const accounts = Object.keys(owners);
  return {
    group_id: groupId,
    nodes: [
      ...unique(Object.values(owners)).map((client) => node('client', client)),
      ...accounts.map((account) => node('account', account)),
      ...unique(logins.map((login) => login.device)).map((device) => node('device', device)),
      ...unique(logins.map((login) => login.address)).map((address) =>
        node('network_address', address)
      ),
    ],
    ownerships: accounts.map((account) => ({
      id: `ownership:${owners[account]}:${account}`,
      source: `client:${owners[account]}`,
      target: `account:${account}`,
    })),
    logins: logins.flatMap((login) => [
      {
        id: `login:${login.event}:device`,
        event_id: login.event,
        occurred_at: login.at,
        source: `account:${login.account}`,
        target: `device:${login.device}`,
      },
      {
        id: `login:${login.event}:address`,
        event_id: login.event,
        occurred_at: login.at,
        source: `account:${login.account}`,
        target: `network_address:${login.address}`,
      },
    ]),
    transfers: transfers.map((transfer) => ({
      id: `transfer:${transfer.event}`,
      event_id: transfer.event,
      occurred_at: transfer.at,
      source: `account:${transfer.from}`,
      target: `account:${transfer.to}`,
      amount: transfer.rub * 100,
    })),
  };
}

const ownedByOwnClients = (accounts: string[]) =>
  Object.fromEntries(accounts.map((account) => [account, `cli-${account}`]));

const farmAt = (time: string) => `2026-11-12T${time}:00Z`;
const farmId = '6f1c2a3e-1b7d-4c1a-9e51-0a1b2c3d4e03';
const farm: MockGroup = {
  item: {
    id: farmId,
    scheme: 'farm',
    status: 'new',
    risk_level: 'high',
    risk_score: 75,
    risk_parts: { farm: { scheme: 50, speed: 15, clients: 0, collector: 10 } },
    detected_at: farmAt('10:30'),
    first_event_at: farmAt('10:01'),
    last_event_at: farmAt('10:29'),
    features: { accounts_count: 4, span_minutes: 28, total_amount: 16100000 },
  },
  graph: buildGraph(farmId, {
    owners: {
      'acc-ravil': 'cli-ravil',
      'acc-sasha': 'cli-sasha',
      'acc-dima': 'cli-dima',
      'acc-collector': 'cli-organizer',
    },
    logins: [
      {
        event: 'evt-df-with-0001',
        at: farmAt('10:01'),
        account: 'acc-ravil',
        device: 'dev-common',
        address: '198.51.100.21',
      },
      {
        event: 'evt-df-with-0002',
        at: farmAt('10:09'),
        account: 'acc-sasha',
        device: 'dev-common',
        address: '198.51.100.22',
      },
      {
        event: 'evt-df-with-0003',
        at: farmAt('10:16'),
        account: 'acc-dima',
        device: 'dev-common',
        address: '198.51.100.23',
      },
    ],
    transfers: [
      {
        event: 'evt-df-with-0004',
        at: farmAt('10:20'),
        from: 'acc-ravil',
        to: 'acc-collector',
        rub: 52000,
      },
      {
        event: 'evt-df-with-0005',
        at: farmAt('10:24'),
        from: 'acc-sasha',
        to: 'acc-collector',
        rub: 42000,
      },
      {
        event: 'evt-df-with-0006',
        at: farmAt('10:29'),
        from: 'acc-dima',
        to: 'acc-collector',
        rub: 67000,
      },
    ],
  }),
};

const at = (time: string) => `2026-10-28T${time}:00Z`;

const addressId = '6f1c2a3e-1b7d-4c1a-9e51-0a1b2c3d4e04';
const address: MockGroup = {
  item: {
    id: addressId,
    scheme: 'address',
    status: 'new',
    risk_level: 'medium',
    risk_score: 70,
    risk_parts: {},
    detected_at: at('13:59'),
    first_event_at: at('13:00'),
    last_event_at: at('13:58'),
    features: { accounts_count: 6, span_minutes: 58, total_amount: 21000000 },
  },
  graph: buildGraph(addressId, {
    owners: ownedByOwnClients(['a-1', 'a-2', 'a-3', 'a-4', 'a-5', 'a-6']),
    logins: [
      {
        event: 'evt-addr-01',
        at: at('13:00'),
        account: 'a-1',
        device: 'pc-01',
        address: '178.4.2.10',
      },
      {
        event: 'evt-addr-02',
        at: at('13:06'),
        account: 'a-2',
        device: 'pc-01',
        address: '178.4.2.10',
      },
      {
        event: 'evt-addr-03',
        at: at('13:12'),
        account: 'a-3',
        device: 'pc-02',
        address: '178.4.2.10',
      },
      {
        event: 'evt-addr-04',
        at: at('13:19'),
        account: 'a-4',
        device: 'phone-9c',
        address: '178.4.2.10',
      },
      {
        event: 'evt-addr-05',
        at: at('13:25'),
        account: 'a-5',
        device: 'phone-9c',
        address: '178.4.2.10',
      },
    ],
    transfers: [
      { event: 'evt-addr-06', at: at('13:30'), from: 'a-1', to: 'a-6', rub: 42000 },
      { event: 'evt-addr-07', at: at('13:38'), from: 'a-2', to: 'a-6', rub: 42000 },
      { event: 'evt-addr-08', at: at('13:44'), from: 'a-3', to: 'a-6', rub: 42000 },
      { event: 'evt-addr-09', at: at('13:51'), from: 'a-4', to: 'a-6', rub: 42000 },
      { event: 'evt-addr-10', at: at('13:58'), from: 'a-5', to: 'a-6', rub: 42000 },
    ],
  }),
};

const chainId = '6f1c2a3e-1b7d-4c1a-9e51-0a1b2c3d4e02';
const chain: MockGroup = {
  item: {
    id: chainId,
    scheme: 'chain',
    status: 'new',
    risk_level: 'high',
    risk_score: 90,
    risk_parts: {},
    detected_at: at('11:28'),
    first_event_at: at('11:02'),
    last_event_at: at('11:27'),
    features: { accounts_count: 5, span_minutes: 25, total_amount: 197000000 },
  },
  graph: buildGraph(chainId, {
    owners: ownedByOwnClients(['c-1', 'c-2', 'c-3', 'c-4', 'c-5']),
    logins: [],
    transfers: [
      { event: 'evt-chain-01', at: at('11:02'), from: 'c-1', to: 'c-2', rub: 500000 },
      { event: 'evt-chain-02', at: at('11:09'), from: 'c-2', to: 'c-3', rub: 495000 },
      { event: 'evt-chain-03', at: at('11:18'), from: 'c-3', to: 'c-4', rub: 490000 },
      { event: 'evt-chain-04', at: at('11:27'), from: 'c-4', to: 'c-5', rub: 485000 },
    ],
  }),
};

const ringId = '6f1c2a3e-1b7d-4c1a-9e51-0a1b2c3d4e01';
const ring: MockGroup = {
  item: {
    id: ringId,
    scheme: 'ring',
    status: 'new',
    risk_level: 'high',
    risk_score: 95,
    risk_parts: {},
    detected_at: at('09:25'),
    first_event_at: at('09:10'),
    last_event_at: at('09:24'),
    features: { accounts_count: 4, span_minutes: 14, total_amount: 237000000 },
  },
  graph: buildGraph(ringId, {
    owners: ownedByOwnClients(['r-1', 'r-2', 'r-3', 'r-4']),
    logins: [],
    transfers: [
      { event: 'evt-ring-01', at: at('09:10'), from: 'r-1', to: 'r-2', rub: 300000 },
      { event: 'evt-ring-02', at: at('09:12'), from: 'r-2', to: 'r-3', rub: 298000 },
      { event: 'evt-ring-03', at: at('09:14'), from: 'r-3', to: 'r-4', rub: 296000 },
      { event: 'evt-ring-04', at: at('09:16'), from: 'r-4', to: 'r-1', rub: 294000 },
      { event: 'evt-ring-05', at: at('09:18'), from: 'r-1', to: 'r-2', rub: 300000 },
      { event: 'evt-ring-06', at: at('09:20'), from: 'r-2', to: 'r-3', rub: 297000 },
      { event: 'evt-ring-07', at: at('09:22'), from: 'r-3', to: 'r-4', rub: 295000 },
      { event: 'evt-ring-08', at: at('09:24'), from: 'r-4', to: 'r-1', rub: 290000 },
    ],
  }),
};

export const mockGroups: MockGroup[] = [farm, address, chain, ring];
