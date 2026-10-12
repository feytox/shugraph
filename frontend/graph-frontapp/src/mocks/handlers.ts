import { http, HttpResponse } from 'msw';
import type { Note, ProblemDetails } from '@/api/client';

let inMemoryNotes: Note[] = [
  {
    id: 'note-1',
    text: 'Тестовая операция: перевод со счета 123 на 456',
    created_at: '2026-10-07T10:00:00.000Z',
  },
  {
    id: 'note-2',
    text: 'Вход с подозрительного IP 192.168.1.1',
    created_at: '2026-10-07T10:05:00.000Z',
  },
];

export const resetMockNotes = () => {
  inMemoryNotes = [
    {
      id: 'note-1',
      text: 'Тестовая операция: перевод со счета 123 на 456',
      created_at: '2026-10-07T10:00:00.000Z',
    },
    {
      id: 'note-2',
      text: 'Вход с подозрительного IP 192.168.1.1',
      created_at: '2026-10-07T10:05:00.000Z',
    },
  ];
};

export const handlers = [
  http.get('*/api/v1/notes', () => {
    return HttpResponse.json({ notes: inMemoryNotes, Notes: inMemoryNotes }, { status: 200 });
  }),

  http.post('*/api/v1/notes', async ({ request }) => {
    const body = (await request.json()) as { text?: string };

    if (!body || !body.text || body.text.trim() === '') {
      return HttpResponse.json<ProblemDetails>(
        {
          title: 'Unprocessable Entity',
          status: 422,
          detail: 'Поле text обязательно для заполнения',
          type: 'about:blank',
          errors: [{ message: 'expected length >= 1', location: 'body.text', value: body?.text }],
        },
        { status: 422, headers: { 'Content-Type': 'application/problem+json' } }
      );
    }

    const newNote: Note = {
      id: `note-${Date.now()}`,
      text: body.text.trim(),
      created_at: new Date().toISOString(),
    };

    inMemoryNotes.push(newNote);
    return HttpResponse.json(newNote, { status: 200 });
  }),
];
