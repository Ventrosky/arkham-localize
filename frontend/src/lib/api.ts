export type SupportedLanguage = 'it' | 'fr' | 'de' | 'es';

export const SUPPORTED_LANGUAGES = [
  { code: 'it', name: 'Italian' },
  { code: 'fr', name: 'French' },
  { code: 'de', name: 'German' },
  { code: 'es', name: 'Spanish' },
] as const;

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

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:3001';

export async function translate(
  text: string,
  language: SupportedLanguage
): Promise<TranslateResponse> {
  const response = await fetch(`${API_URL}/translate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ text, language }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(error.error || `HTTP error! status: ${response.status}`);
  }

  return response.json();
}
