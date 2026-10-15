import { http, HttpResponse } from 'msw';
import type { Note } from '@/api/client';
import type { operations } from '@/api/v1';

type ListNotesResponse =
  operations['get-all-notes']['responses'][200]['content']['application/json'];
type CreateNoteBody = operations['create-note']['requestBody']['content']['application/json'];
type CreateNoteResponse =
  operations['create-note']['responses'][201]['content']['application/json'];
type CreateNoteError =
  operations['create-note']['responses']['default']['content']['application/problem+json'];

const validationError = (message: string, location: string, value: unknown) =>
  HttpResponse.json<CreateNoteError>(
    {
      type: 'about:blank',
      title: 'Unprocessable Entity',
      status: 422,
      detail: 'validation failed',
      errors: [{ message, location, value }],
    },
    { status: 422, headers: { 'Content-Type': 'application/problem+json' } }
  );

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
  http.get<never, never, ListNotesResponse>('*/api/v1/notes', () => {
    return HttpResponse.json<ListNotesResponse>({ notes: inMemoryNotes }, { status: 200 });
  }),

  http.post<never, CreateNoteBody, CreateNoteResponse | CreateNoteError>(
    '*/api/v1/notes',
    async ({ request }) => {
      let body: Partial<CreateNoteBody> | null;
      try {
        body = (await request.json()) as Partial<CreateNoteBody> | null;
      } catch {
        return validationError('request body is required', 'body', null);
      }

      if (typeof body?.text !== 'string') {
        return validationError('expected required property text to be present', 'body', body);
      }
      if (body.text.length < 1) {
        return validationError('expected length >= 1', 'body.text', body.text);
      }

      const newNote: Note = {
        id: crypto.randomUUID(),
        text: body.text,
        created_at: new Date().toISOString(),
      };

      inMemoryNotes.push(newNote);
      return HttpResponse.json<CreateNoteResponse>(newNote, { status: 201 });
    }
  ),
];
