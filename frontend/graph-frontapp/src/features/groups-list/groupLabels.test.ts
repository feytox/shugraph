import { describe, expect, it } from 'vitest';
import { accountsLabel } from '@/features/groups-list/groupLabels';

describe('accountsLabel', () => {
  it.each([
    [1, '1 счёт'],
    [2, '2 счёта'],
    [4, '4 счёта'],
    [5, '5 счетов'],
    [11, '11 счетов'],
    [12, '12 счетов'],
    [21, '21 счёт'],
    [22, '22 счёта'],
    [100, '100 счетов'],
  ])('%i → %s', (count, label) => {
    expect(accountsLabel(count)).toBe(label);
  });
});
