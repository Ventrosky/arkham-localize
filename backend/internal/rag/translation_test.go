package rag

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateTranslation_SimilarToMachete(t *testing.T) {
	// Skip if OPENAI_API_KEY is not set
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test (set OPENAI_API_KEY to enable)")
	}

	// Use text very similar to Machete for testing (with more variation)
	// Machete: "[action]: <b>Fight.</b> You get +1 [combat] for this attack. If the attacked enemy is the only enemy engaged with you, this attack deals +1 damage."
	// Modified: Changed +1 to +2, and condition from "if only enemy" to "if enemy is NOT the only one" (opposite condition)
	englishText := "[action]: <b>Fight.</b> You get +2 [combat] for this attack. If this enemy is not the only one engaged with you, this attack deals +1 damage."

	// Create context cards similar to Machete (melee weapons)
	contextCards := []ContextCard{
		{
			CardName:    "Machete",
			CardCode:    "01020",
			IsBack:      false,
			EnglishText: "[action]: <b>Fight.</b> You get +1 [combat] for this attack. If the attacked enemy is the only enemy engaged with you, this attack deals +1 damage.",
			ItalianText: "[action]: <b>Combatti.</b> Ricevi +1 [combat] in questo attacco. Se il nemico attaccato è l'unico nemico ingaggiato con te, questo attacco infligge +1 danno.",
		},
		{
			CardName:    "Survival Knife",
			CardCode:    "03003",
			IsBack:      false,
			EnglishText: "[action]: <b>Fight.</b> You get +1 [combat] for this attack.",
			ItalianText: "[action]: <b>Combatti.</b> Ricevi +1 [combat] in questo attacco.",
		},
	}

	// Generate translation
	translation, err := GenerateTranslation(englishText, contextCards, apiKey)
	if err != nil {
		t.Fatalf("Failed to generate translation: %v", err)
	}

	if translation == "" {
		t.Fatal("Translation is empty")
	}

	// Remove quotes if the LLM wrapped the response in quotes
	cleanTranslation := strings.Trim(translation, `"`)

	t.Logf("Original: %s", englishText)
	t.Logf("Translation: %s", cleanTranslation)

	// Verify that the translation starts correctly
	expectedStart := "[action]:"
	if !strings.HasPrefix(cleanTranslation, expectedStart) {
		t.Errorf("Translation should start with '%s', got: %s", expectedStart, cleanTranslation)
	}

	// Verify that symbols are preserved
	testCases := []struct {
		symbol string
		desc   string
	}{
		{"[action]", "action symbol"},
		{"[combat]", "combat symbol"},
		{"<b>", "HTML bold tag"},
		{"</b>", "HTML bold closing tag"},
		{"+2", "mathematical symbol (+2 should be preserved)"},
		{"+1", "mathematical symbol (+1 should be preserved)"},
	}

	for _, tc := range testCases {
		if !strings.Contains(cleanTranslation, tc.symbol) {
			t.Errorf("Translation should preserve %s (%s), but it's missing. Translation: %s", tc.symbol, tc.desc, cleanTranslation)
		}
	}

	// Verify that Italian words are present (not just English)
	// The translation should contain Italian words like "Combatti" or "Ricevi"
	italianKeywords := []string{"Combatti", "Ricevi", "attacco", "danno", "nemico"}
	foundItalian := false
	for _, keyword := range italianKeywords {
		if strings.Contains(cleanTranslation, keyword) {
			foundItalian = true
			break
		}
	}

	if !foundItalian {
		t.Logf("Warning: Translation might not be in Italian. Translation: %s", cleanTranslation)
		// Don't fail, just warn - the LLM might use different wording
	}

	// Verify that the translation is similar to the context (Machete translation)
	// It should be close to: "[action]: <b>Combatti.</b> Ricevi +1 [combat] in questo attacco..."
	if !strings.Contains(cleanTranslation, "Combatti") || !strings.Contains(cleanTranslation, "Ricevi") {
		t.Logf("Warning: Translation might not match context style. Expected similar to Machete translation.")
	}
}
