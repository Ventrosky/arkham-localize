import { ContextCard } from '../lib/api'

interface ContextCardsProps {
  cards: ContextCard[]
}

interface CardContextDisplayProps {
  card: ContextCard
}

function CardContextDisplay({ card }: CardContextDisplayProps) {
  const arkhamDbUrl = card.card_code 
    ? `https://arkhamdb.com/card/${card.card_code}` 
    : null;

  return (
    <div className="p-3 border border-gray-700 bg-gray-800 rounded-lg shadow-inner flex flex-col gap-1">
      {arkhamDbUrl ? (
        <a 
          href={arkhamDbUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="text-lg font-bold text-yellow-300 hover:text-yellow-200 hover:underline transition-colors cursor-pointer"
        >
          {card.card_name}
        </a>
      ) : (
        <h4 className="text-lg font-bold text-yellow-300">{card.card_name}</h4>
      )}
      {card.card_code && (
        <div className="flex items-center gap-2 mb-2">
          {arkhamDbUrl ? (
            <a
              href={arkhamDbUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="bg-gray-700 px-2 py-1 rounded text-xs text-gray-300 hover:bg-gray-600 hover:text-white transition-colors cursor-pointer"
            >
              {card.card_code}
            </a>
          ) : (
            <span className="bg-gray-700 px-2 py-1 rounded text-xs text-gray-300">
              {card.card_code}
            </span>
          )}
          {card.is_back && (
            <span className="bg-blue-900 text-blue-200 px-2 py-1 rounded text-xs">
              Back
            </span>
          )}
        </div>
      )}
      <p className="text-sm font-semibold text-gray-400">EN:</p>
      <p className="text-sm italic text-gray-200 whitespace-pre-wrap">{card.english_text}</p>
      <p className="text-sm font-semibold text-gray-400 mt-2">IT:</p>
      <p className="text-sm italic text-gray-200 whitespace-pre-wrap">{card.italian_text}</p>
    </div>
  )
}

export default function ContextCards({ cards }: ContextCardsProps) {
  if (cards.length === 0) {
    return null
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-3 text-yellow-200 border-b border-gray-700 pb-2">
        Grounded Context (Vector Search Matches)
      </h2>
      <p className="text-gray-400 text-sm mb-4">
        These are the existing English/Italian card pairs retrieved from the Vector Database based on similarity to your input. This context is sent to the AI to ensure consistent LCG terminology.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {cards.map((card, index) => (
          <CardContextDisplay key={`${card.card_code}-${index}`} card={card} />
        ))}
      </div>
    </div>
  )
}
