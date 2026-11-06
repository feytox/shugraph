import { act, render, type RenderResult } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { vi } from 'vitest';
import { App } from '@/App';

export function renderScreen(): RenderResult {
  return render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <App />
    </QueryClientProvider>
  );
}

export async function advance(milliseconds: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(milliseconds);
  });
}

export async function waitForScreen(assertion: () => void) {
  await act(async () => {
    await vi.waitFor(assertion);
  });
}
