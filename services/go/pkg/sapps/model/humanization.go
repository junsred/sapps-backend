package model

import "time"

type Humanization struct {
	ID                string     `json:"id" db:"id"`
	InputText         string     `json:"input_text" db:"input_text"`
	OutputText        *string    `json:"output_text" db:"output_text"`
	OriginalDetection *float64   `json:"original_detection" db:"original_detection"`
	FinalDetection    *float64   `json:"final_detection" db:"final_detection"`
	ImprovementScore  *float64   `json:"improvement_score" db:"improvement_score"`
	Status            string     `json:"status" db:"status"` // "processing", "completed", "failed"
	CreatedDate       time.Time  `json:"created_date" db:"created_date"`
	ProcessedDate     *time.Time `json:"processed_date" db:"processed_date"`
	UserID            string     `json:"user_id" db:"user_id"`
}
