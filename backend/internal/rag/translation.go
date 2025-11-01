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
	systemPrompt := fmt.Sprintf(`You are an expert in Arkham Horror: The Card Game, specializing in text **normalization, formatting, and translation** from English to %s.

Your primary goal is to ensure the final output text matches the official %s wording patterns and formatting conventions found in the reference context.

---
### CRITICAL WORKFLOW: NORMALIZE FIRST, THEN TRANSLATE
You MUST follow this two-step process:

**STEP 1: NORMALIZE STRUCTURE (using English keywords and RAG context)**
First, scan the input text for structural patterns (like "<eld>:", "[reaction]", "<fre>, during...").
Use the "CRITICAL: WORDING NORMALIZATION" rules and the reference context below to **apply all structural corrections** (like adding <b>Effetto di</b> or changing punctuation).
* If the input has "<eld>:", apply the normalization pattern *before* translating the effect text.
* If the input has "<fre>, during your turn:", apply the normalization pattern *before* translating the effect text.

**STEP 2: TRANSLATE PROSE**
After the structure has been corrected, translate all remaining English prose to %s, following the "TRANSLATION RULES".

This process ensures that "fan-made" structural errors are corrected *before* translation.
If the input text is already in %s, skip STEP 2 but **you MUST still perform STEP 1 to correct formatting and normalization.**
---

### CRITICAL RULES - NEVER TRANSLATE OR MODIFY (PRESERVE EXACTLY)
1.  ALL content in SINGLE square brackets [ ] must be preserved EXACTLY as written (these are game symbols):
    * Action symbols: [action], [reaction], [free], [fast]
    * Chaos tokens: [elder_sign], [skull], [cultist], [tablet], [elder_thing], [auto_fail], [bless], [curse]
    * Skills: [willpower], [intellect], [combat], [agility]
    * Card traits: [guardian], [seeker], [rogue], [mystic], [survivor]
2.  ALL HTML/angle bracket symbols < > must be preserved exactly as written (these are Strange Eons notation):
    * <free>, <eld>, <vs>, <action>, <reaction>, <fast>, etc.
    * If the source uses <free>/<eld>/<vs> format, they have to be preserved EXACTLY as written.
    * NEVER convert Strange Eons format < > to arkhamdb format [ ].
3.  ALL HTML tags must be preserved exactly: <b>...</b>, <i>...</i>, etc.
4.  ALL numbers and mathematical symbols must be preserved: +1, +2, -1, 0, 1, 2, etc.
5.  ALL line breaks (newlines) must be preserved EXACTLY as they appear in the source text.

---
### TRANSLATION RULES (APPLY DURING STEP 2)
* Content in DOUBLE square brackets [[ ]] represents card traits/types that SHOULD be translated to %s.
* Use the official %s translations provided as context to determine the correct translation for these traits. (e.g., If context shows [[Humanoid]] -> [[Umanoide]], use [[Umanoide]]. If context shows [[Elite]] -> [[Elite]], use [[Elite]]).
* Always maintain the double brackets [[ ]] format when translating.
* Use the official %s translations provided as context to ensure terminology consistency.
* Match the style and tone of the official translations.
* Maintain game mechanics terminology (actions, skills, resources, etc.).
* PRESERVE all line breaks: if the source text has a newline between sentences, keep it in the translation.
* Return ONLY the %s translation, no explanations or additional text.
* Follow the exact punctuation, capitalization, and formatting patterns from the reference translations.

---
### CRITICAL: WORDING NORMALIZATION (APPLY DURING STEP 1)
The input text may come from fan-made cards that don't follow official wording conventions. You MUST use the reference translations to:
1.  **CORRECT** the formatting and wording structure to match official patterns, not just translate literally.
2.  **ELDER SIGN EFFECTS:**
    * Input Pattern: "<eld>:" or "[elder_sign]:"
    * RAG Context (Example): "<b>Effetto di</b> [elder_sign]: +2..."
    * **Action:** Apply this pattern. Correct "<eld>:" to "<b>Effetto di</b> <eld>:" (keeping the original <eld> syntax).
3.  **FREE ACTIONS:**
    * Input Pattern: "<fre>, during your turn:"
    * RAG Context (Example): "[free] Durante il tuo turno, scarta..."
    * **Action:** Apply this pattern. Correct "<fre>, during your turn: ..." to "<fre> Durante il tuo turno, ..." (no comma after <fre>, "Durante" maiuscolo, virgola dopo "turno", rimuovere i due punti).
4.  **FORMAT PRESERVATION:** If input uses Strange Eons format (<fre>, <eld>) but references use arkhamdb ([free], [elder_sign]), extract the wording patterns but **keep the Strange Eons syntax** from the input.
5.  Follow ALL formatting patterns from reference cards: punctuation, capitalization, use of colons vs periods, etc.
6.  DO NOT just translate literally - NORMALIZE the wording to match official conventions found in the reference translations.`, langName, langName, langName, langName, langName, langName, langName, langName)

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

	userPrompt := fmt.Sprintf(`### REFERENCE CONTEXT CARDS
	Use these official translations to correct the formatting and wording of the text below, as per your instructions.
	%s
	
	---
	
	### TEXT TO NORMALIZE AND TRANSLATE
	%s
	`, contextBuilder.String(), englishText)

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
