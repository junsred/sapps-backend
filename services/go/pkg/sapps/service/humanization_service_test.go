package service

import (
	"context"
	"sapps/lib/connection"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanizeText(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test.")
	}

	chatGPT, err := connection.NewChatGPT()
	assert.NoError(t, err)

	humanizationService := NewHumanizationService(chatGPT)

	inputText := `1.6 Summary of Methodology

This study will adopt a mixed-methods approach, combining quantitative surveys with qualitative interviews. A structured questionnaire will be used to gather data on the extent of mHealth adoption, influencing factors, and perceived benefits from pharmacy staff across selected pharmacies in Accra. In-depth interviews will be conducted with key informants such as pharmacy managers and health tech developers to gain deeper insights into the barriers and enablers of mHealth adoption.

The sample will include both independent and chain pharmacies to ensure a diverse perspective. Data will be analyzed using statistical software (e.g., SPSS) for the quantitative aspect, and thematic analysis will be employed for qualitative data.`
	options := Options{
		Tone:               "professional",
		FixGrammar:         true,
		SentenceComplexity: "medium",
	}

	result, err := humanizationService.HumanizeText(context.Background(), inputText, options)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.HumanizedText)
	assert.NotEqual(t, inputText, result.HumanizedText)
	assert.NotEmpty(t, result.OriginalDetection)
	assert.NotEmpty(t, result.FinalDetection)
	t.Log(result.HumanizedText)
	t.Log(result.OriginalDetection)
	t.Log(result.FinalDetection)
	t.Log(result.ImprovementScore)
}
