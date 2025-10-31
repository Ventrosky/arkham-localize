package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GenerateTranslation generates a translation using GPT-4o
// with context from similar cards
// language is one of: "it", "fr", "de", "es"
func GenerateTranslation(englishText string, contextCards []ContextCard, apiKey string, language string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// Map language codes to full names
	langNames := map[string]string{
		"it": "Italian",
		"fr": "French",
		"de": "German",
		"es": "Spanish",
	}
	langName := langNames[language]
	if langName == "" {
		langName = language // Fallback
	}

	// Build system prompt with instructions
	systemPrompt := fmt.Sprintf(`You are an expert translator specializing in Arkham Horror: The Card Game translations from English to %s.

CRITICAL RULES - NEVER TRANSLATE OR MODIFY:
1. ALL content in SINGLE square brackets [ ] must be preserved EXACTLY as written (these are game symbols):
   - Action symbols: [action], [reaction], [free], [fast]
   - Chaos tokens: [elder_sign], [skull], [cultist], [tablet], [elder_thing], [auto_fail], [bless], [curse]
   - Skills: [willpower], [intellect], [combat], [agility]
   - Card traits: [guardian], [seeker], [rogue], [mystic], [survivor]
   - Modifiers: [per_investigator], [per_location], [per_enemy]
   - Other: [seal_a], [seal_b], etc.
2. ALL HTML tags must be preserved exactly: <b>...</b>, <i>...</i>, etc.
3. ALL numbers and mathematical symbols must be preserved: +1, +2, -1, 0, 1, 2, etc.
4. ALL punctuation and formatting (parentheses, colons, periods) must be preserved
5. ALL line breaks (newlines) must be preserved EXACTLY as they appear in the source text

TRANSLATION RULES:
- Content in DOUBLE square brackets [[ ]] represents card traits/types that SHOULD be translated to %s
- However, keep technical terms like [[Elite]] or [[Extradimensional]] in English if they are not commonly translated.
- Always maintain the double brackets [[ ]] format when translating.
- Use the official %s translations provided as context to ensure terminology consistency
- Match the style and tone of the official translations
- Maintain game mechanics terminology (actions, skills, resources, etc.)
- PRESERVE all line breaks: if the source text has a newline between sentences, keep it in the translation
- Return ONLY the %s translation, no explanations or additional text

REMEMBER: Single brackets [ ] are game symbols (DO NOT TRANSLATE). Double brackets [[ ]] are traits/types (TRANSLATE but keep [[ ]] format). Line breaks must be preserved exactly.`, langName, langName, langName, langName)

	// Build user prompt with context
	var contextBuilder strings.Builder
	if len(contextCards) > 0 {
		contextBuilder.WriteString(fmt.Sprintf("Official %s card translations for reference:\n\n", langName))
		for i, card := range contextCards {
			contextBuilder.WriteString(fmt.Sprintf("Card %d: %s (%s)\n", i+1, card.CardName, card.CardCode))
			contextBuilder.WriteString(fmt.Sprintf("English: %s\n", card.EnglishText))
			contextBuilder.WriteString(fmt.Sprintf("%s: %s\n\n", langName, card.TranslatedText))
		}
	}

	userPrompt := fmt.Sprintf(`%sTranslate the following English text to %s:

    "%s"

    CRITICAL TRANSLATION RULES:
    - Do NOT translate anything inside SINGLE square brackets [ ] - these are game symbols that must remain EXACTLY as written (e.g., [action], [elder_sign], [willpower])
    - DO translate content inside DOUBLE square brackets [[ ]] to %s, but maintain the [[ ]] format
    - PRESERVE all line breaks (newlines) exactly as they appear in the source text above
    - Only translate the words outside of single brackets [ ].`,
		contextBuilder.String(), langName, englishText, langName)

	reqBody := struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Temperature float64   `json:"temperature"`
	}{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3, // Lower temperature for more consistent translations
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no translation returned")
	}

	translation := strings.TrimSpace(result.Choices[0].Message.Content)
	return translation, nil
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
