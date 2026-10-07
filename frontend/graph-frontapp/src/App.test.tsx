import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { server } from '@/mocks/server';
import { App } from '@/App';

function renderWithClient() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  );
}

describe('NotesApp (GSD-5)', () => {
  it('отображает заголовок и загружает список записей через MSW', async () => {
    renderWithClient();

    expect(screen.getByText('Тестовая страница (GSD-5 / B1)')).toBeInTheDocument();

    await waitFor(() => {
      expect(
        screen.getByText('Тестовая операция: перевод со счета 123 на 456')
      ).toBeInTheDocument();
      expect(screen.getByText('Вход с подозрительного IP 192.168.1.1')).toBeInTheDocument();
    });
  });

  it('создает новую запись и обновляет список', async () => {
    renderWithClient();

    await waitFor(() => {
      expect(screen.getByText('Тестовая операция: перевод со счета 123 на 456')).toBeInTheDocument();
    });

    const input = screen.getByLabelText('Текст записи');
    const submitButton = screen.getByRole('button', { name: 'Создать' });

    fireEvent.change(input, { target: { value: 'Новая подозрительная цепочка переводов' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Новая подозрительная цепочка переводов')).toBeInTheDocument();
    });
  });

  it('показывает сообщение об ошибке, если бэкенд возвращает ошибку', async () => {
    server.use(
      http.get('*/api/v1/notes', () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    renderWithClient();

    await waitFor(() => {
      expect(screen.getByText('Ошибка подключения')).toBeInTheDocument();
      expect(
        screen.getByText('Не удалось загрузить список записей с сервера')
      ).toBeInTheDocument();
    });
  });
});
