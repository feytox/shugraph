import createClient from 'openapi-fetch';
import type { paths, components } from './v1';

const getBaseUrl = () => {
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin;
  }
  return 'http://localhost:8080';
};

export const apiClient = createClient<paths>({
  baseUrl: getBaseUrl(),
  fetch: (...args: Parameters<typeof globalThis.fetch>) => globalThis.fetch(...args),
});

export type Note = components['schemas']['Note'];
export type CreateNoteRequest = components['schemas']['CreateNoteRequest'];
