import { useState } from 'react';
import TranslationForm from './components/TranslationForm';
import TranslationResult from './components/TranslationResult';
import ContextCards from './components/ContextCards';
import { translate, TranslateResponse, SupportedLanguage } from './lib/api';
import './App.css';

function App() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<TranslateResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleTranslate = async (text: string, language: SupportedLanguage) => {
    if (!text.trim()) {
      setError('Please enter text to translate');
      return;
    }

    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const response = await translate(text, language);
      setResult(response);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to translate');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 p-4 sm:p-8 font-['Inter'] flex flex-col">
      <div className="max-w-4xl mx-auto ah-bg p-6 rounded-xl shadow-2xl border border-gray-700 flex-1">
        <header className="text-center mb-8">
          <h1 className="text-4xl font-extrabold text-yellow-500 mb-2">Arkham Localize</h1>
          <p className="text-gray-400 text-lg">Arkham Horror LCG Consistent Content Translator</p>
        </header>

        {/* Input Section */}
        <div className="mb-8">
          <TranslationForm onSubmit={handleTranslate} loading={loading} />
        </div>

        {/* Error Section */}
        {error && (
          <div className="mb-8 p-4 rounded-lg border border-red-600 bg-red-900/20 text-red-200">
            {error}
          </div>
        )}

        {/* Output Section */}
        {result && (
          <div className="mb-10">
            <TranslationResult translation={result.translation} loading={loading} />
          </div>
        )}

        {/* Context Section */}
        {result && result.context.length > 0 && (
          <div>
            <ContextCards cards={result.context} />
          </div>
        )}
      </div>
      {/* Footer */}
      <footer className="mt-auto py-4 text-center w-full max-w-4xl mx-auto px-4 sm:px-8">
        <p className="text-xs text-gray-500">
          The information presented on this site about{' '}
          <a
            href="https://www.fantasyflightgames.com/en/products/arkham-horror-the-card-game/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-gray-400 hover:text-gray-300 underline"
          >
            Arkham Horror: The Card Game™
          </a>{' '}
          is copyrighted by Fantasy Flight Games.
          <br />
          This website is not produced, endorsed, supported, or affiliated with Fantasy Flight
          Games.
        </p>
      </footer>
    </div>
  );
}

export default App;
