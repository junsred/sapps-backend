package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sapps/lib/connection"
	"sapps/pkg/sapps/model"
	"math"
	"math/rand"
	"regexp"
	"strings"

	"github.com/openai/openai-go"
)

type DetectionService struct {
	chatGPT *connection.ChatGPT
}

type DetectionResult struct {
	AIProbability  float64                  `json:"ai_probability"`
	Breakdown      model.DetectionBreakdown `json:"breakdown"`
	ServiceResults model.ServiceResults     `json:"service_results"`
	DetectionUsage openai.CompletionUsage   `json:"-"`
}

func NewDetectionService(chatGPT *connection.ChatGPT) *DetectionService {
	return &DetectionService{
		chatGPT: chatGPT,
	}
}

func (d *DetectionService) DetectAIContent(ctx context.Context, inputText string) (*DetectionResult, error) {
	// Get detailed sentence analysis
	sentenceAnalysis, usage, err := d.analyzeSentences(ctx, inputText)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze sentences: %w", err)
	}

	// Calculate breakdown statistics
	breakdown := d.calculateBreakdown(sentenceAnalysis)

	// Generate mock service results (simulating various AI detection services)
	serviceResults := d.generateServiceResults(breakdown.AISentences, breakdown.HumanSentences)

	// Calculate overall AI probability
	aiProbability := float64(breakdown.AISentences) / float64(breakdown.AISentences+breakdown.HumanSentences)

	return &DetectionResult{
		AIProbability:  aiProbability,
		Breakdown:      breakdown,
		ServiceResults: serviceResults,
		DetectionUsage: usage,
	}, nil
}

func (d *DetectionService) analyzeSentences(ctx context.Context, text string) ([]model.SentenceAnalysis, openai.CompletionUsage, error) {
	systemPrompt := `You are an expert AI text classifier. Your task is to analyze the provided text sentence by sentence and determine the probability that each sentence was written by AI.

For each sentence, analyze these indicators:
1. **Perplexity & Predictability**: Generic or overly predictable phrasing
2. **Burstiness**: Monotonous sentence structure vs natural rhythm variation
3. **Voice**: Sterile, neutral tone vs personal, distinctive voice
4. **Grammar Perfection**: Unnaturally perfect grammar vs natural imperfections
5. **Word Choice**: Generic vocabulary vs specific, vivid language
6. **Flow**: Artificial transitions vs natural conversation flow

For each sentence, provide a probability score between 0.0 and 1.0 where:
- 0.0-0.2: Very likely human-written
- 0.2-0.4: Likely human-written  
- 0.4-0.6: Uncertain/mixed
- 0.6-0.8: Likely AI-written
- 0.8-1.0: Very likely AI-written

Respond with a JSON object containing sentence-by-sentence analysis.`

	userPrompt := fmt.Sprintf("Analyze this text sentence by sentence for AI detection:\n\n%s", text)

	type sentenceDetectionResponse struct {
		SentenceAnalysis []model.SentenceAnalysis `json:"sentence_analysis"`
	}

	response, usage, err := d.chatGPT.GenerateCompletion(ctx, openai.ChatModelGPT4_1Mini, systemPrompt, userPrompt, openai.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
			JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "sentence_ai_detection",
				Description: openai.String("Sentence-by-sentence AI detection analysis"),
				Strict:      openai.Bool(true),
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"sentence_analysis": map[string]interface{}{
							"type":        "array",
							"description": "Analysis of each sentence",
							"items": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"text": map[string]interface{}{
										"type":        "string",
										"description": "The sentence text",
									},
									"ai_probability": map[string]interface{}{
										"type":        "number",
										"description": "Probability that this sentence is AI-generated (0.0-1.0)",
										"minimum":     0.0,
										"maximum":     1.0,
									},
								},
								"required":             []string{"text", "ai_probability"},
								"additionalProperties": false,
							},
						},
					},
					"required":             []string{"sentence_analysis"},
					"additionalProperties": false,
				},
			},
		},
	}, 0.3)
	if err != nil {
		return nil, openai.CompletionUsage{}, fmt.Errorf("failed to get sentence analysis: %w", err)
	}

	var parsedResponse sentenceDetectionResponse
	if err := json.Unmarshal([]byte(response), &parsedResponse); err != nil {
		return nil, openai.CompletionUsage{}, fmt.Errorf("failed to parse sentence analysis JSON: %w", err)
	}

	// If no sentences were analyzed, try to split manually
	if len(parsedResponse.SentenceAnalysis) == 0 {
		sentences := d.splitIntoSentences(text)
		for _, sentence := range sentences {
			parsedResponse.SentenceAnalysis = append(parsedResponse.SentenceAnalysis, model.SentenceAnalysis{
				Text:          sentence,
				AIProbability: 0.5, // Default uncertain score
			})
		}
	}

	return parsedResponse.SentenceAnalysis, usage, nil
}

func (d *DetectionService) splitIntoSentences(text string) []string {
	// Simple sentence splitting regex
	re := regexp.MustCompile(`[.!?]+\s+`)
	sentences := re.Split(text, -1)

	var result []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}

	if len(result) == 0 {
		result = append(result, strings.TrimSpace(text))
	}

	return result
}

func (d *DetectionService) calculateBreakdown(sentenceAnalysis []model.SentenceAnalysis) model.DetectionBreakdown {
	humanSentences := 0
	aiSentences := 0
	totalConfidence := 0.0

	for _, analysis := range sentenceAnalysis {
		if analysis.AIProbability < 0.5 {
			humanSentences++
		} else {
			aiSentences++
		}
		// Convert probability to confidence (higher probability = higher confidence in classification)
		confidence := math.Max(analysis.AIProbability, 1.0-analysis.AIProbability)
		totalConfidence += confidence
	}

	avgConfidence := 0.0
	if len(sentenceAnalysis) > 0 {
		avgConfidence = totalConfidence / float64(len(sentenceAnalysis))
	}

	return model.DetectionBreakdown{
		HumanSentences:   humanSentences,
		AISentences:      aiSentences,
		Confidence:       avgConfidence,
		SentenceAnalysis: sentenceAnalysis,
	}
}

func (d *DetectionService) generateServiceResults(aiSentences, humanSentences int) model.ServiceResults {
	// Calculate base AI probability
	total := aiSentences + humanSentences
	if total == 0 {
		total = 1
	}
	baseAIProbability := float64(aiSentences) / float64(total)
	if baseAIProbability < 0.1 {
		return model.ServiceResults{
			GPTZero:   0.0,
			Writer:    0.0,
			QuillBot:  0.0,
			Copyleaks: 0.0,
			Sapling:   0.0,
		}
	}
	// Add some realistic variation around the base probability
	return model.ServiceResults{
		GPTZero:   d.addVariation(baseAIProbability, 0.15),
		Writer:    d.addVariation(1.0-baseAIProbability, 0.2), // Writer tends to be more human-friendly
		QuillBot:  d.addVariation(baseAIProbability, 0.25),
		Copyleaks: d.addVariation(baseAIProbability, 0.12),
		Sapling:   d.addVariation(baseAIProbability, 0.18),
		Grammarly: d.addVariation(1.0-baseAIProbability, 0.15), // Grammarly tends to be more human-friendly
		ZeroGPT:   d.addVariation(baseAIProbability, 0.1),
	}
}

func (d *DetectionService) addVariation(base float64, variance float64) float64 {
	// Add random variation within the specified range
	variation := (rand.Float64() - 0.5) * 2 * variance
	result := base + variation

	// Clamp to [0, 1] range
	if result < 0.0 {
		result = 0.0
	} else if result > 1.0 {
		result = 1.0
	}

	return math.Round(result*100) / 100 // Round to 2 decimal places
}
