import { apiClient } from '@/api/client';

export async function listScenarios(signal: AbortSignal) {
  const { data } = await apiClient.GET('/admin-api/v1/scenarios', { signal });

  if (!data) {
    throw new Error('Не удалось загрузить сценарии.');
  }

  return data;
}

export async function runScenario(scenarioId: string) {
  const { response } = await apiClient
    .POST('/admin-api/v1/scenarios/{scenario_id}/run', {
      params: { path: { scenario_id: scenarioId } },
    })
    .catch(() => {
      throw new Error('Ответ на запрос не получен. Проверьте соединение.');
    });

  if (response.status === 404) {
    throw new Error('Сценарий больше недоступен.');
  }

  if (response.status !== 202) {
    throw new Error(`Сервер не принял запуск (HTTP ${response.status}).`);
  }
}
