import type { AskCitation, AskDonePayload } from '../types';

export interface AskCallbacks {
  onCitations?: (citations: AskCitation[]) => void;
  onDelta?: (text: string) => void;
  onDone?: (payload: AskDonePayload) => void;
  onError?: (message: string) => void;
}

// askQuestion POSTs to /api/ask and parses the SSE stream.
// Uses fetch (not axios) because axios buffers response bodies in the browser.
export async function askQuestion(
  question: string,
  includeFullText: boolean,
  callbacks: AskCallbacks,
  signal?: AbortSignal
): Promise<void> {
  let res: Response;
  try {
    res = await fetch('/api/ask', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question, include_full_text: includeFullText }),
      signal,
    });
  } catch (err) {
    if ((err as Error).name === 'AbortError') return;
    callbacks.onError?.('Couldn\'t reach the server');
    return;
  }

  if (!res.ok) {
    let msg = 'Couldn\'t reach the model';
    try {
      const body = await res.json();
      if (body.error) msg = body.error;
    } catch {
      // keep default message
    }
    callbacks.onError?.(msg);
    return;
  }

  if (!res.body) {
    callbacks.onError?.('Streaming not supported in this browser');
    return;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  const dispatch = (rawEvent: string) => {
    let event = 'message';
    let data = '';
    for (const line of rawEvent.split('\n')) {
      if (line.startsWith('event: ')) event = line.slice(7).trim();
      else if (line.startsWith('data: ')) data += line.slice(6);
    }
    if (!data) return;

    let parsed: unknown;
    try {
      parsed = JSON.parse(data);
    } catch {
      return;
    }

    switch (event) {
      case 'citations':
        callbacks.onCitations?.(parsed as AskCitation[]);
        break;
      case 'delta':
        callbacks.onDelta?.((parsed as { text: string }).text);
        break;
      case 'done':
        callbacks.onDone?.(parsed as AskDonePayload);
        break;
      case 'error':
        callbacks.onError?.((parsed as { error: string }).error);
        break;
    }
  };

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      let idx: number;
      while ((idx = buffer.indexOf('\n\n')) !== -1) {
        const rawEvent = buffer.slice(0, idx);
        buffer = buffer.slice(idx + 2);
        dispatch(rawEvent);
      }
    }
    if (buffer.trim()) dispatch(buffer);
  } catch (err) {
    if ((err as Error).name !== 'AbortError') {
      callbacks.onError?.('Connection interrupted');
    }
  }
}
