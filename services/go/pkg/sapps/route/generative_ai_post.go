package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/dig"
)

type PostGenerativeAI struct {
	dig.In
	MainDB *maindb.MainDB
}

type PostGenerativeAIRequest struct {
	ImageID string `json:"image_id"`
	Prompt  string `json:"prompt"`
}

type KieCreateTaskRequest struct {
	Model       string             `json:"model"`
	CallBackURL string             `json:"callBackUrl"`
	Input       KieCreateTaskInput `json:"input"`
}

type KieCreateTaskInput struct {
	ImageUrls    []string `json:"image_urls"`
	Prompt       string   `json:"prompt"`
	OutputFormat string   `json:"output_format"`
	ImageSize    string   `json:"image_size"`
}

type KieCreateTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID string `json:"taskId"`
	} `json:"data"`
}

type PostGenerativeAIResponse struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
}

func (r *PostGenerativeAI) Handler(c *middleware.RequestContext) error {
	if c.User().PremiumType == nil {
		return c.Error(middleware.NewStatus(fiber.StatusBadRequest, "SHOW_PAYWALL"))
	}

	var req PostGenerativeAIRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Error(middleware.StatusBadRequest, "invalid request body")
	}

	if req.ImageID == "" {
		return c.Error(middleware.StatusBadRequest, "image_id is required")
	}
	if req.Prompt == "" {
		return c.Error(middleware.StatusBadRequest, "prompt is required")
	}

	var existingTask PostGenerativeAIResponse
	err := r.MainDB.QueryRow(c.Context(),
		`SELECT id, task_id, status, EXTRACT(EPOCH FROM created_at)::bigint 
		 FROM generative_ai_tasks 
		 WHERE user_id = $1 AND image_id = $2 AND LOWER(prompt) = LOWER($3)
		 LIMIT 1`,
		c.UserID(), req.ImageID, req.Prompt).Scan(&existingTask.ID, &existingTask.TaskID, &existingTask.Status, &existingTask.CreatedAt)

	if err == nil {
		return c.JSON(existingTask)
	}

	imageURL := fmt.Sprintf("%s/cdn/img/%s.jpg", constant.API_URL, req.ImageID)

	kieReq := KieCreateTaskRequest{
		Model:       "google/nano-banana-edit",
		CallBackURL: fmt.Sprintf("%s/webhook/kie/callback", constant.API_URL),
		Input: KieCreateTaskInput{
			ImageUrls:    []string{imageURL},
			Prompt:       req.Prompt,
			OutputFormat: "jpeg",
			ImageSize:    "auto",
		},
	}

	kieReqBody, err := json.Marshal(kieReq)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to create request")
	}

	httpReq, err := http.NewRequest("POST", "https://api.kie.ai/api/v1/jobs/createTask", bytes.NewBuffer(kieReqBody))
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+constant.GetKiaAPIKey())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to send request to AI service")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to read response")
	}

	var kieResp KieCreateTaskResponse
	if err := json.Unmarshal(body, &kieResp); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to parse response")
	}

	if kieResp.Code != 200 {
		c.LogErr(fmt.Errorf("kie api error: %s %d", kieResp.Message, kieResp.Code))
		return c.Error(middleware.StatusInternalServerError, "AI service error")
	}

	id := uuid.New().String()
	createdAt := time.Now()

	_, err = r.MainDB.Exec(c.Context(),
		`INSERT INTO generative_ai_tasks (id, user_id, image_id, prompt, task_id, status, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, c.UserID(), req.ImageID, req.Prompt, kieResp.Data.TaskID, "pending", createdAt)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to save task")
	}

	return c.JSON(PostGenerativeAIResponse{
		ID:        id,
		TaskID:    kieResp.Data.TaskID,
		Status:    "pending",
		CreatedAt: createdAt.Unix(),
	})
}
