package route

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sapps/lib/connection"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	"go.uber.org/dig"
)

type PostScan struct {
	dig.In
	MainDB  *maindb.MainDB
	ChatGPT *connection.ChatGPT
}

type PostScanRequest struct {
	ImageID string `json:"image_id"`
}

type ScanData struct {
	FaceOverallRating int `json:"face_overall_rating"`
	FemininityRating  int `json:"femininity_rating"`
	MasculinityRating int `json:"masculinity_rating"`
	FaceRating        int `json:"face_rating"`
	EyesRating        int `json:"eyes_rating"`
	JawlineRating     int `json:"jawline_rating"`
	SkinRating        int `json:"skin_rating"`
}

type PostScanResponse struct {
	ScanID    string   `json:"scan_id"`
	Data      ScanData `json:"data"`
	CreatedAt int64    `json:"created_at"`
}

const scanSystemPrompt = `You are an expert facial analysis AI. Analyze the provided face image and rate the following features on a scale of 1-10.

Respond ONLY in TOML format with the following structure:
face_overall_rating = "X/10"
femininity_rating = "X/10"
masculinity_rating = "X/10"
face_rating = "X/10"
eyes_rating = "X/10"
jawline_rating = "X/10"
skin_rating = "X/10"

Replace X with your rating number. Be honest and objective in your assessment.`

func parseRating(s string) int {
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}
	val, _ := strconv.Atoi(match)
	return val
}

func parseTOMLResponse(response string) ScanData {
	data := ScanData{}
	re := regexp.MustCompile(`(\w+)\s*=\s*"?(\d+)/?10?"?`)
	matches := re.FindAllStringSubmatch(response, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		key := match[1]
		val, _ := strconv.Atoi(match[2])
		switch key {
		case "face_overall_rating":
			data.FaceOverallRating = val
		case "femininity_rating":
			data.FemininityRating = val
		case "masculinity_rating":
			data.MasculinityRating = val
		case "face_rating":
			data.FaceRating = val
		case "eyes_rating":
			data.EyesRating = val
		case "jawline_rating":
			data.JawlineRating = val
		case "skin_rating":
			data.SkinRating = val
		}
	}
	return data
}

func (r *PostScan) Handler(c *middleware.RequestContext) error {
	var req PostScanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Error(middleware.StatusBadRequest, "invalid request body")
	}

	if req.ImageID == "" {
		return c.Error(middleware.StatusBadRequest, "image_id is required")
	}

	imageURL := fmt.Sprintf("%s/cdn/img/%s.jpg", constant.API_URL, req.ImageID)

	response, _, err := r.ChatGPT.GenerateCompletionWithImage(
		context.Background(),
		shared.ChatModelGPT5,
		scanSystemPrompt,
		imageURL,
		openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &openai.ResponseFormatTextParam{},
		},
		1,
	)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to analyze image")
	}
	log.Println(response)

	scanData := parseTOMLResponse(response)
	scanDataJSON, err := json.Marshal(scanData)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to process scan data")
	}

	scanID := uuid.New().String()
	createdAt := time.Now()

	_, err = r.MainDB.Exec(c.Context(),
		"INSERT INTO scans (scan_id, user_id, image_id, data, created_at) VALUES ($1, $2, $3, $4, $5)",
		scanID, c.UserID(), req.ImageID, scanDataJSON, createdAt)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to save scan")
	}

	return c.JSON(PostScanResponse{
		ScanID:    scanID,
		Data:      scanData,
		CreatedAt: createdAt.Unix(),
	})
}
