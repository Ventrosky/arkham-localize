import { useState } from 'react'
import TranslationForm from './components/TranslationForm'
import TranslationResult from './components/TranslationResult'
import ContextCards from './components/ContextCards'
import { translate, TranslateResponse } from './lib/api'
import './App.css'

function App() {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<TranslateResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleTranslate = async (text: string) => {
    if (!text.trim()) {
      setError('Please enter text to translate')
      return
    }

    setLoading(true)
    setError(null)
    setResult(null)

    try {
      const response = await translate(text)
      setResult(response)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to translate')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-900 text-gray-100 p-4 sm:p-8 font-['Inter']">
      <div className="max-w-4xl mx-auto ah-bg p-6 rounded-xl shadow-2xl border border-gray-700">
        <header className="text-center mb-8">
          <h1 className="text-4xl font-extrabold text-yellow-500 mb-2">
            Arkham Horror LCG Agentic Translator
          </h1>
          <p className="text-gray-400 text-lg">
            Grounded translation using vector similarity for consistent terminology.
          </p>
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
      <footer className="mt-12 text-center">
        <p className="text-xs text-gray-500">
          <a
            href="https://www.fantasyflightgames.com/en/products/arkham-horror-the-card-game/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-gray-400 hover:text-gray-300 underline"
          >
            Arkham Horror: The Card Game™
          </a>
          {' '}and all related content © Fantasy Flight Games (FFG). This site is not produced, endorsed by or affiliated with FFG.
        </p>
      </footer>
    </div>
  )
}

export default App
