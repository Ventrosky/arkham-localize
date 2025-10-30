export interface ContextCard {
  card_name: string;
  card_code: string;
  is_back: boolean;
  english_text: string;
  italian_text: string;
}

export interface TranslateResponse {
  translation: string;
  context: ContextCard[];
}

export interface TranslateRequest {
  text: string;
}

const API_BASE_URL = (import.meta.env?.VITE_API_URL as string) || 'http://localhost:3001';

export async function translate(text: string): Promise<TranslateResponse> {
  const response = await fetch(`${API_BASE_URL}/translate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ text }),
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(`Translation failed: ${error}`);
  }

  return response.json();
}

export async function healthCheck(): Promise<{ status: string; service: string }> {
  const response = await fetch(`${API_BASE_URL}/health`);
  if (!response.ok) {
    throw new Error('Health check failed');
  }
  return response.json();
}

