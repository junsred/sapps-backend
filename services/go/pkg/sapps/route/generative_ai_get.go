package route

import (
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type GetGenerativeAI struct {
	dig.In
	MainDB *maindb.MainDB
}

type GetGenerativeAIResponse struct {
	ID          string  `json:"id"`
	TaskID      string  `json:"task_id"`
	ImageID     string  `json:"image_id"`
	Prompt      string  `json:"prompt"`
	Status      string  `json:"status"`
	ResultURL   *string `json:"result_url,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	CompletedAt *int64  `json:"completed_at,omitempty"`
}

func (r *GetGenerativeAI) Handler(c *middleware.RequestContext) error {
	taskID := c.Params("id")
	if taskID == "" {
		return c.Error(middleware.StatusBadRequest, "id is required")
	}

	var resp GetGenerativeAIResponse
	var completedAt *int64

	row := r.MainDB.QueryRow(c.Context(),
		`SELECT id, task_id, image_id, prompt, status, result_url, 
		        EXTRACT(EPOCH FROM created_at)::bigint,
		        EXTRACT(EPOCH FROM completed_at)::bigint
		 FROM generative_ai_tasks 
		 WHERE (id = $1 OR task_id = $1) AND user_id = $2`,
		taskID, c.UserID())

	err := row.Scan(&resp.ID, &resp.TaskID, &resp.ImageID, &resp.Prompt, &resp.Status, &resp.ResultURL, &resp.CreatedAt, &completedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Error(middleware.NewStatus(fiber.StatusNotFound, "NOT_FOUND"), "task not found")
		}
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to fetch task")
	}

	resp.CompletedAt = completedAt

	return c.JSON(resp)
}
