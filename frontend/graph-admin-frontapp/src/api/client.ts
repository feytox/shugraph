import createClient from 'openapi-fetch';
import type { paths } from './v1';

export const apiClient = createClient<paths>({ baseUrl: window.location.origin });
