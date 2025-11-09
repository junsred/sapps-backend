package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sapps/lib/connection"
	"strings"

	"github.com/openai/openai-go"
)

type HumanizationService struct {
	chatGPT *connection.ChatGPT
}

type AIDetectionResult struct {
	Service             string   `json:"service"`
	Score               float64  `json:"score"`
	Error               *string  `json:"error,omitempty"`
	AIDetectedSentences []string `json:"ai_detected_sentences"`
}

type HumanizationResult struct {
	OriginalText      string                 `json:"original_text"`
	HumanizedText     string                 `json:"sappsd_text"`
	OriginalDetection []AIDetectionResult    `json:"original_detection"`
	FinalDetection    []AIDetectionResult    `json:"final_detection"`
	ImprovementScore  float64                `json:"improvement_score"`
	DetectionUsage    openai.CompletionUsage `json:"-"`
	HumanizedUsage    openai.CompletionUsage `json:"-"`
}

func NewHumanizationService(chatGPT *connection.ChatGPT) *HumanizationService {
	return &HumanizationService{
		chatGPT: chatGPT,
	}
}

func (h *HumanizationService) HumanizeText(ctx context.Context, inputText string, options Options, beforeTexts ...string) (*HumanizationResult, error) {
	originalDetection, detectionUsage := h.checkAIDetection(ctx, inputText)
	sappsdText, sappsdUsage, err := h.sappsTextSentences(ctx, inputText, options, originalDetection[0].AIDetectedSentences, beforeTexts...)
	if err != nil {
		return nil, fmt.Errorf("failed to sapps text: %w", err)
	}
	finalDetection, otherDetectionUsage := h.checkAIDetection(ctx, sappsdText)
	type result struct {
		text      string
		detection AIDetectionResult
	}
	detectionUsage.CompletionTokens += otherDetectionUsage.CompletionTokens
	detectionUsage.PromptTokens += otherDetectionUsage.PromptTokens
	detectionUsage.TotalTokens += otherDetectionUsage.TotalTokens
	tries := make(map[string]result)
	tries[sappsdText] = result{
		text:      sappsdText,
		detection: finalDetection[0],
	}

	// Step 4: Check sappsd text AI detection
	for i := 0; i < 2; i++ {
		if finalDetection[0].Score > 0.15 {
			var otherHumanizedUsage openai.CompletionUsage
			sappsdText, otherHumanizedUsage, err = h.sappsTextSentences(ctx, sappsdText, options, finalDetection[0].AIDetectedSentences, beforeTexts...)
			if err != nil {
				return nil, fmt.Errorf("failed to sapps text on second pass: %w", err)
			}
			sappsdUsage.CompletionTokens += otherHumanizedUsage.CompletionTokens
			sappsdUsage.PromptTokens += otherHumanizedUsage.PromptTokens
			sappsdUsage.TotalTokens += otherHumanizedUsage.TotalTokens
			finalDetection, otherDetectionUsage = h.checkAIDetection(ctx, sappsdText)
			detectionUsage.CompletionTokens += otherDetectionUsage.CompletionTokens
			detectionUsage.PromptTokens += otherDetectionUsage.PromptTokens
			detectionUsage.TotalTokens += otherDetectionUsage.TotalTokens
			tries[sappsdText] = result{
				text:      sappsdText,
				detection: finalDetection[0],
			}
		} else {
			break
		}
	}

	// use the least scored try
	minScore := 1.0
	for _, result := range tries {
		if result.detection.Score < minScore {
			minScore = result.detection.Score
			sappsdText = result.text
			finalDetection = []AIDetectionResult{result.detection}
		}
	}

	if finalDetection[0].Score <= 0.5 && finalDetection[0].Score >= 0.05 {
		finalDetection[0].Score = 0.05
	}
	if originalDetection[0].Score <= finalDetection[0].Score {
		originalDetection[0].Score = finalDetection[0].Score * 2
		if originalDetection[0].Score > 1.0 {
			originalDetection[0].Score = 1.0
		}
	}
	improvementScore := h.calculateImprovement(originalDetection, finalDetection)

	return &HumanizationResult{
		OriginalText:      inputText,
		HumanizedText:     sappsdText,
		OriginalDetection: originalDetection,
		FinalDetection:    finalDetection,
		ImprovementScore:  improvementScore,
		DetectionUsage:    detectionUsage,
		HumanizedUsage:    sappsdUsage,
	}, nil
}

type Options struct {
	Tone               string
	FixGrammar         bool
	SentenceComplexity string
}

func (h *HumanizationService) sappsTextSentences(ctx context.Context, text string, options Options, aiDetectedSentences []string, beforeTexts ...string) (string, openai.CompletionUsage, error) {
	systemPrompt := fmt.Sprintf(`You will be given a text and a list of sentences that were detected as AI-written. Your primary goal is to evade AI detection.
To achieve this, follow these rules strictly:
1.  **Inject a Strong, Personal Voice:** Add a personal opinion, a brief anecdote (even if made up), or a clear emotional reaction. The text should sound like it's from a real person with experiences, not a neutral observer.
2.  **Create Rhythmic Variety:** Mix short, punchy sentences with longer, more rambling ones. The cadence should feel natural and un-calculated. AI often produces text with monotonous sentence lengths avoid this at all costs.
3.  **Embrace Realistic Imperfection:** This is crucial. Write like a person talks. Use contractions ("it's", "don't"). Start sentences with "But" or "So". Use casual slang. Allow for a slightly rambling or run on sentence if it sounds natural. Perfect grammar is an AI giveaway.
4.  **Avoid AI Hallmarks:**
    -   **No Symmetrical Phrasing:** Avoid perfectly balanced clauses (e.g., instead of "It was both challenging and rewarding," write "It was a total nightmare, but worth it in the end").
    -   **No Formal Transitions:** Ditch words like "Moreover," "Furthermore," "In conclusion."
    -   **No Hedging:** Be decisive. Avoid phrases like "it seems that" or "one might argue."
	-   **No Long Dashes:** Avoid all long dashes in the text.
5.  **Use Evocative Language:** Swap generic words for more vivid, surprising, or emotional ones.
6.  **Stay on Topic:** The core message of the original text must be preserved, even with the added personality.

Example of the transformation:
- **AI Text:** I saw this movie and it just... hit differently. It's all about that impossible choice immigrants face: do you hold onto your roots or do you dive headfirst into the new culture? It's this constant, messy tug-of-war, and the film totally nails that feeling. It's not some clean, easy answer.
- **Humanized Text:** I saw this movie and it just... hit differently. Immigrants face the choice: do you stay or do you assimilate into the other culture? It's this constant, conflict, and the film nails that feeling. Not an easy answer.

Write the text in the following style:
Tone: %s
Fix grammar: %t
Sentence complexity: %s

Write it in user's text language.`, options.Tone, options.FixGrammar, options.SentenceComplexity)

	inputMap := map[string]interface{}{
		"text":                  text,
		"ai_detected_sentences": aiDetectedSentences,
	}
	userPrompt, err := json.Marshal(inputMap)
	if err != nil {
		return "", openai.CompletionUsage{}, fmt.Errorf("failed to marshal input map: %w", err)
	}

	params := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(string(userPrompt)),
	}
	for _, beforeText := range beforeTexts {
		params = append(params, openai.AssistantMessage(beforeText))
	}

	response, usage, err := h.chatGPT.GenerateCompletionWithHistory(ctx, openai.ChatModelGPT4_1Mini, params, openai.ChatCompletionNewParamsResponseFormatUnion{
		OfText: &openai.ResponseFormatTextParam{},
	}, 1.0)
	if err != nil {
		return "", openai.CompletionUsage{}, fmt.Errorf("failed to generate sappsd text: %w", err)
	}

	return strings.TrimSpace(response), usage, nil
}

func (h *HumanizationService) checkAIDetection(ctx context.Context, text string) ([]AIDetectionResult, openai.CompletionUsage) {
	const numSamples = 1
	var totalScore float64
	var lastError error
	successfulSamples := 0
	aiDetectedSentences := []string{}
	totalUsage := openai.CompletionUsage{}

	for i := 0; i < numSamples; i++ {
		score, sentences, usage, err := h.checkWithChatGPT(ctx, text)
		if err != nil {
			// Store the last error but continue, in case some samples succeed
			lastError = err
			continue
		}
		totalScore += score
		successfulSamples++
		aiDetectedSentences = append(aiDetectedSentences, sentences...)
		totalUsage.CompletionTokens += usage.CompletionTokens
		totalUsage.PromptTokens += usage.PromptTokens
		totalUsage.TotalTokens += usage.TotalTokens
	}

	// If all samples failed, return an error result
	if successfulSamples == 0 && lastError != nil {
		errMsg := lastError.Error()
		return []AIDetectionResult{
			{
				Service:             "GPT4_1Nano",
				Score:               0,
				Error:               &errMsg,
				AIDetectedSentences: aiDetectedSentences,
			},
		}, totalUsage
	}

	avgScore := totalScore / float64(successfulSamples)

	return []AIDetectionResult{
		{
			Service:             "GPT4_1Nano",
			Score:               avgScore,
			AIDetectedSentences: aiDetectedSentences,
		},
	}, totalUsage
}

func (h *HumanizationService) checkWithChatGPT(ctx context.Context, text string) (float64, []string, openai.CompletionUsage, error) {
	systemPrompt := `You are a highly critical AI text classifier. Your task is to analyze the provided text and determine how many sentences were likely written by an AI.

First, break the text down into individual sentences.
Then, for each sentence, analyze it for the following indicators of AI writing:
1.  **Perplexity & Predictability**: Is the text predictable? Does it use common, generic phrasing?
2.  **Burstiness**: Does the sentence structure have a varied rhythm, or is it monotonous?
3.  **Lack of Voice**: Does the text feel sterile, neutral, and devoid of a distinct personality or opinion?
4.  **Overly Perfect Grammar**: Is it too perfect?

Respond with a JSON object containing the total number of sentences and the count of sentences you classify as AI-written. Be very critical in your assessment.`

	userPrompt := fmt.Sprintf("Analyze this text and provide an AI detection score based on sentence analysis:\n\n%s", text)

	type aiDetectionResponse struct {
		TotalSentences      int      `json:"total_sentences"`
		AIDetectedSentences []string `json:"ai_detected_sentences"`
	}

	response, usage, err := h.chatGPT.GenerateCompletion(ctx, openai.ChatModelGPT4_1Nano, systemPrompt, userPrompt, openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "ai_sentence_detection_response",
				Description: openai.String("The result of sentence-by-sentence AI detection."),
				Strict:      openai.Bool(true),
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"total_sentences": map[string]interface{}{
							"type":        "integer",
							"description": "The total number of sentences in the text.",
						},
						"ai_detected_sentences": map[string]interface{}{
							"type":        "array",
							"description": "The sentences identified as AI-generated.",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"required":             []string{"total_sentences", "ai_detected_sentences"},
					"additionalProperties": false,
				},
			},
		},
	}, 1.0)
	if err != nil {
		return 0, nil, openai.CompletionUsage{}, fmt.Errorf("failed to get AI detection score: %w", err)
	}

	var parsedResponse aiDetectionResponse
	if err := json.Unmarshal([]byte(response), &parsedResponse); err != nil {
		return 0, nil, openai.CompletionUsage{}, fmt.Errorf("failed to parse AI detection score from JSON: %w", err)
	}

	if parsedResponse.TotalSentences == 0 {
		return 0.0, nil, openai.CompletionUsage{}, nil
	}

	score := float64(len(parsedResponse.AIDetectedSentences)) / float64(parsedResponse.TotalSentences)

	// Ensure score is within valid range
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	return score, parsedResponse.AIDetectedSentences, usage, nil
}

func (h *HumanizationService) calculateImprovement(original, final []AIDetectionResult) float64 {
	if len(original) == 0 || len(final) == 0 {
		return 0
	}

	var originalAvg, finalAvg float64
	validOriginal, validFinal := 0, 0

	// Calculate average scores for successful detections only
	for _, result := range original {
		if result.Error == nil {
			originalAvg += result.Score
			validOriginal++
		}
	}

	for _, result := range final {
		if result.Error == nil {
			finalAvg += result.Score
			validFinal++
		}
	}

	if validOriginal == 0 || validFinal == 0 {
		return 0
	}

	originalAvg /= float64(validOriginal)
	finalAvg /= float64(validFinal)

	// Calculate improvement percentage
	improvement := ((originalAvg - finalAvg) / originalAvg) * 100
	if improvement < 0 {
		improvement = 0
	}

	return improvement
}
