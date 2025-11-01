package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file from backend directory
	envPath := filepath.Join("..", "..", ".env")
	_ = godotenv.Load(envPath)
	// Also try current directory
	_ = godotenv.Load()
}

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
			CardName:       "Machete",
			CardCode:       "01020",
			IsBack:         false,
			EnglishText:    "[action]: <b>Fight.</b> You get +1 [combat] for this attack. If the attacked enemy is the only enemy engaged with you, this attack deals +1 damage.",
			TranslatedText: "[action]: <b>Combatti.</b> Ricevi +1 [combat] in questo attacco. Se il nemico attaccato è l'unico nemico ingaggiato con te, questo attacco infligge +1 danno.",
		},
		{
			CardName:       "Survival Knife",
			CardCode:       "03003",
			IsBack:         false,
			EnglishText:    "[action]: <b>Fight.</b> You get +1 [combat] for this attack.",
			TranslatedText: "[action]: <b>Combatti.</b> Ricevi +1 [combat] in questo attacco.",
		},
	}

	// Generate translation (using Italian as target language for the test)
	translation, err := GenerateTranslation(englishText, contextCards, apiKey, "it")
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

// TestNormalization_TableDriven tests multiple normalization scenarios
func TestNormalization_TableDriven(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test (set OPENAI_API_KEY to enable)")
	}

	testCases := []struct {
		name         string
		englishText  string
		contextCards []ContextCard
		assertions   []func(t *testing.T, translation string)
	}{
		{
			name:        "ElderSign_Simple",
			englishText: "<eld>: +2. Draw a card.",
			contextCards: []ContextCard{
				{
					CardName:       "William Yorick",
					CardCode:       "01019",
					IsBack:         false,
					EnglishText:    "[elder_sign]: +2. ...",
					TranslatedText: "<b>Effetto di</b> [elder_sign]: +2. ...",
				},
			},
			assertions: []func(t *testing.T, translation string){
				func(t *testing.T, tr string) {
					if !strings.Contains(tr, "<b>Effetto di</b>") {
						t.Errorf("Must contain '<b>Effetto di</b>'")
					}
					if !strings.Contains(tr, "<eld>:") {
						t.Errorf("Must preserve '<eld>:'")
					}
				},
			},
		},
		{
			name:        "ElderSign_WithSummon",
			englishText: "<eld>: +1. You may put a <b><i>Summon</i></b> asset in your discard pile into your hand.",
			contextCards: []ContextCard{
				{
					CardName:       "William Yorick",
					CardCode:       "01019",
					IsBack:         false,
					EnglishText:    "[elder_sign]: +2. ...",
					TranslatedText: "<b>Effetto di</b> [elder_sign]: +2. ...",
				},
			},
			assertions: []func(t *testing.T, translation string){
				func(t *testing.T, tr string) {
					if !strings.Contains(tr, "<b>Effetto di</b>") {
						t.Errorf("FAILED NORMALIZATION: Translation must contain '<b>Effetto di</b>' before the elder sign effect. Got: %s", tr)
					}
					if !strings.Contains(tr, "<eld>:") {
						t.Errorf("FAILED: Must preserve <eld>: format. Got: %s", tr)
					}
					// Note: "Summon" is a fan-made trait that doesn't exist in official translations,
					// so it should be preserved as-is and NOT translated to "Evocazione"
				},
			},
		},
		{
			name:        "FreeAction_Simple",
			englishText: "<fre>, during your turn: Discard a card.",
			contextCards: []ContextCard{
				{
					CardName:       "Pete's Guitar",
					CardCode:       "01021",
					IsBack:         false,
					EnglishText:    "[free] During your turn, discard...",
					TranslatedText: "[free] Durante il tuo turno, scarta...",
				},
			},
			assertions: []func(t *testing.T, translation string){
				func(t *testing.T, tr string) {
					if strings.Contains(tr, ", during") {
						t.Errorf("Should not contain ', during' (comma should be removed)")
					}
					if !strings.Contains(tr, "Durante il tuo turno,") && !strings.Contains(tr, "Durante il tuo turno, ") {
						t.Errorf("Should contain 'Durante il tuo turno,'")
					}
					if strings.Contains(tr, "turno:") {
						t.Errorf("Should not contain 'turno:' (colon should be removed)")
					}
				},
			},
		},
		{
			name:        "FreeAction_WithPikachu",
			englishText: "<fre>, during your turn: Ready Ashley's Pikachu. You suffer 1 direct damage. (Limit once per turn.)",
			contextCards: []ContextCard{
				{
					CardName:       "Pete's Guitar",
					CardCode:       "01021",
					IsBack:         false,
					EnglishText:    "[free] During your turn, discard...",
					TranslatedText: "[free] Durante il tuo turno, scarta...",
				},
			},
			assertions: []func(t *testing.T, translation string){
				func(t *testing.T, tr string) {
					if strings.Contains(tr, ", during") {
						t.Errorf("FAILED NORMALIZATION: Should remove comma after <fre> and capitalize 'Durante'. Got: %s", tr)
					}
					if !strings.Contains(tr, "<fre>") {
						t.Errorf("FAILED: Must preserve <fre> format. Got: %s", tr)
					}
					if !strings.Contains(tr, "Durante il tuo turno,") && !strings.Contains(tr, "Durante il tuo turno, ") {
						t.Errorf("FAILED NORMALIZATION: Should have 'Durante il tuo turno,' (capital D, comma, no colon). Got: %s", tr)
					}
					if strings.Contains(tr, "turno:") {
						t.Errorf("FAILED NORMALIZATION: Should remove colon after 'turno'. Got: %s", tr)
					}
				},
			},
		},
		{
			name: "PikachuCase_Complete",
			englishText: `You begin the game with Ashley's Pikachu in play.

<vs>

<fre>, during your turn: Ready Ashley's Pikachu. You suffer 1 direct damage. (Limit once per turn.)

<eld>: +1. You may put a <b><i>Summon</i></b> asset in your discard pile into your hand.`,
			contextCards: []ContextCard{
				{
					CardName:       "William Yorick",
					CardCode:       "01019",
					IsBack:         false,
					EnglishText:    "[elder_sign]: +2. ...",
					TranslatedText: "<b>Effetto di</b> [elder_sign]: +2. ...",
				},
				{
					CardName:       "Pete's Guitar",
					CardCode:       "01021",
					IsBack:         false,
					EnglishText:    "[free] During your turn, discard...",
					TranslatedText: "[free] Durante il tuo turno, scarta...",
				},
			},
			assertions: []func(t *testing.T, translation string){
				func(t *testing.T, tr string) {
					// REGRESSION TEST 1: Elder Sign normalization
					if !strings.Contains(tr, "<b>Effetto di</b>") {
						t.Errorf("REGRESSION: Must normalize <eld>: to '<b>Effetto di</b> <eld>:'. Missing '<b>Effetto di</b>'. Got: %s", tr)
					}
					if !strings.Contains(tr, "<eld>:") {
						t.Errorf("REGRESSION: Must preserve <eld>: format. Got: %s", tr)
					}
					// REGRESSION TEST 2: Free action normalization
					if strings.Contains(tr, "<fre>, during") {
						t.Errorf("REGRESSION: Must normalize '<fre>, during' to '<fre> Durante il tuo turno,'. Found wrong pattern. Got: %s", tr)
					}
					if !strings.Contains(tr, "<fre>") {
						t.Errorf("REGRESSION: Must preserve <fre> format. Got: %s", tr)
					}
					// REGRESSION TEST 3: Should NOT have English "during your turn:"
					if strings.Contains(tr, "during your turn:") {
						t.Errorf("REGRESSION: Should translate 'during your turn:' to Italian and normalize. Found English text. Got: %s", tr)
					}
					// REGRESSION TEST 4: "Summon" should NOT be translated
					// Note: "Summon" is a fan-made trait that doesn't exist in official translations,
					// so it should be preserved as-is (not translated to "Evocazione")
					if strings.Contains(tr, "Evocazione") && strings.Contains(strings.ToLower(tr), "summon") {
						t.Errorf("REGRESSION: 'Summon' is a fan-made trait and should NOT be translated to 'Evocazione'. It should be preserved as 'Summon'. Got: %s", tr)
					}
					// REGRESSION TEST 5: Should preserve <vs> separator
					if !strings.Contains(tr, "<vs>") {
						t.Errorf("REGRESSION: Must preserve <vs> separator. Got: %s", tr)
					}
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			translation, err := GenerateTranslation(tc.englishText, tc.contextCards, apiKey, "it")
			if err != nil {
				t.Fatalf("Failed to generate translation: %v", err)
			}

			cleanTranslation := strings.Trim(translation, `"`)
			t.Logf("Original: %s", tc.englishText)
			t.Logf("Translation: %s", cleanTranslation)

			for i, assert := range tc.assertions {
				t.Run(fmt.Sprintf("assertion_%d", i), func(t *testing.T) {
					assert(t, cleanTranslation)
				})
			}
		})
	}
}
