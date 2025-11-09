package model

import "time"

type SentenceAnalysis struct {
	Text          string  `json:"text"`
	AIProbability float64 `json:"ai_probability"`
}

type DetectionBreakdown struct {
	HumanSentences   int                `json:"human_sentences"`
	AISentences      int                `json:"ai_sentences"`
	Confidence       float64            `json:"confidence"`
	SentenceAnalysis []SentenceAnalysis `json:"sentence_analysis"`
}

type ServiceResults struct {
	GPTZero   float64 `json:"GPTZero"`
	Writer    float64 `json:"Writer"`
	QuillBot  float64 `json:"QuillBot"`
	Copyleaks float64 `json:"Copyleaks"`
	Sapling   float64 `json:"Sapling"`
	Grammarly float64 `json:"Grammarly"`
	ZeroGPT   float64 `json:"ZeroGPT"`
}

type DetectionResponse struct {
	AIProbability  float64            `json:"ai_probability"`
	Breakdown      DetectionBreakdown `json:"breakdown"`
	ServiceResults ServiceResults     `json:"service_results"`
}

type Detection struct {
	ID             string     `json:"id" db:"id"`
	InputText      string     `json:"input_text" db:"input_text"`
	AIProbability  float64    `json:"ai_probability" db:"ai_probability"`
	BreakdownJSON  string     `json:"-" db:"breakdown_json"`
	ServiceResults string     `json:"-" db:"service_results"`
	Status         string     `json:"status" db:"status"` // "processing", "completed", "failed"
	CreatedDate    time.Time  `json:"created_date" db:"created_date"`
	ProcessedDate  *time.Time `json:"processed_date" db:"processed_date"`
	UserID         string     `json:"user_id" db:"user_id"`
}
