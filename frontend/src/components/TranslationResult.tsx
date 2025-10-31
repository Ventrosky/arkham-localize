interface TranslationResultProps {
  translation: string
  loading: boolean
}

// Removes double quotes from the beginning and end of the translation
function cleanTranslation(translation: string): string {
  return translation.trim().replace(/^"+|"+$/g, '')
}

export default function TranslationResult({ translation, loading }: TranslationResultProps) {
  const cleanedTranslation = translation ? cleanTranslation(translation) : ''
  
  return (
    <div>
      <h2 className="text-xl font-semibold mb-3 text-yellow-200 border-b border-gray-700 pb-2">
        Translation Result
      </h2>
      <div className={`p-4 min-h-[100px] rounded-lg border-2 ${
        cleanedTranslation ? 'border-green-600 bg-gray-800' : 'border-dashed border-gray-700 bg-gray-900'
      } transition-all duration-300`}>
        {loading && (
          <div className="flex items-center justify-center h-full">
            <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-yellow-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span className="text-yellow-500">Awaiting LLM Response...</span>
          </div>
        )}
        {!loading && cleanedTranslation && (
          <p className="text-lg text-white whitespace-pre-wrap">{cleanedTranslation}</p>
        )}
        {!loading && !cleanedTranslation && (
          <p className="text-gray-500 italic">Translation will appear here.</p>
        )}
      </div>
    </div>
  )
}
