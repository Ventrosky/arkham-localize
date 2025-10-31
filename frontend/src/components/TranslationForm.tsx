import { useState, FormEvent } from 'react'
import { SUPPORTED_LANGUAGES, SupportedLanguage } from '../lib/api'

interface TranslationFormProps {
  onSubmit: (text: string, language: SupportedLanguage) => void
  loading: boolean
}

export default function TranslationForm({ onSubmit, loading }: TranslationFormProps) {
  const [text, setText] = useState('')
  const [language, setLanguage] = useState<SupportedLanguage>('it')

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    onSubmit(text, language)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (!loading && text.trim()) {
        onSubmit(text, language)
      }
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <div className="mb-4">
        <label htmlFor="languageSelect" className="block text-xl font-semibold mb-2 text-yellow-200">
          Target Language
        </label>
        <select
          id="languageSelect"
          value={language}
          onChange={(e) => setLanguage(e.target.value as SupportedLanguage)}
          disabled={loading}
          className="w-full sm:w-auto px-4 py-2 rounded-lg bg-gray-800 border border-gray-700 focus:ring-yellow-500 focus:border-yellow-500 text-lg text-gray-100"
        >
          {SUPPORTED_LANGUAGES.map((lang) => (
            <option key={lang.code} value={lang.code}>
              {lang.name}
            </option>
          ))}
        </select>
      </div>
      <label htmlFor="inputText" className="block text-xl font-semibold mb-2 text-yellow-200">
        English Text to Translate
      </label>
      <textarea
        id="inputText"
        rows={4}
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Enter card text here (e.g., '[action] Investigate. If you succeed, deal 1 damage to an enemy at your location.')"
        className="w-full p-4 rounded-lg bg-gray-800 border border-gray-700 focus:ring-yellow-500 focus:border-yellow-500 text-lg text-gray-100"
        disabled={loading}
      />
      <button
        onClick={handleSubmit}
        disabled={loading || !text.trim()}
        className="ah-button mt-4 w-full sm:w-auto px-6 py-3 text-lg font-bold text-white rounded-xl shadow-lg hover:shadow-xl transition duration-300"
      >
        {loading ? 'Translating...' : 'Translate Card Text'}
      </button>
    </form>
  )
}
