export interface ContextCard {
  card_name: string;
  card_code: string;
  is_back: boolean;
  english_text: string;
  translated_text: string;
}

export interface TranslateResponse {
  translation: string;
  context: ContextCard[];
}

export interface TranslateRequest {
  text: string;
  language: string; // "it", "fr", "de", "es"
}

export type SupportedLanguage = 'it' | 'fr' | 'de' | 'es';

export const SUPPORTED_LANGUAGES: { code: SupportedLanguage; name: string }[] = [
  { code: 'it', name: 'Italian' },
  { code: 'fr', name: 'French' },
  { code: 'de', name: 'German' },
  { code: 'es', name: 'Spanish' },
];

const API_BASE_URL = (import.meta.env?.VITE_API_URL as string) || 'http://localhost:3001';

export async function translate(text: string, language: SupportedLanguage = 'it'): Promise<TranslateResponse> {
  const response = await fetch(`${API_BASE_URL}/translate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ text, language }),
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

