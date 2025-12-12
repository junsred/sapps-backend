package route

import (
	"fmt"
	"os"
	"path"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type DeleteGenerativeAI struct {
	dig.In
	MainDB *maindb.MainDB
}

func (r *DeleteGenerativeAI) Handler(c *middleware.RequestContext) error {
	id := c.Params("id")
	if id == "" {
		return c.Error(middleware.StatusBadRequest, "id is required")
	}

	var resultURL *string
	// Check if task exists and belongs to user
	err := r.MainDB.QueryRow(c.Context(),
		"SELECT result_url FROM generative_ai_tasks WHERE id = $1 AND user_id = $2",
		id, c.UserID()).Scan(&resultURL)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return c.Error(middleware.NewStatus(fiber.StatusNotFound, "NOT_FOUND"), "task not found")
		}
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to fetch task")
	}

	// Delete from DB
	_, err = r.MainDB.Exec(c.Context(),
		"DELETE FROM generative_ai_tasks WHERE id = $1 AND user_id = $2",
		id, c.UserID())
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to delete task")
	}

	// Delete file from disk if exists
	if resultURL != nil && *resultURL != "" {
		// Extract filename from URL (assuming format .../cdn/img/filename.jpg)
		filename := path.Base(*resultURL)
		if filename != "" && !strings.Contains(filename, "/") && !strings.Contains(filename, "\\") {
			filePath := fmt.Sprintf("%s/cdn/img/%s", constant.WD_PATH, filename)
			if err := os.Remove(filePath); err != nil {
				// Log error but don't fail request since DB delete was successful
				c.LogErr(fmt.Errorf("failed to delete file %s: %v", filePath, err))
			}
		}
	}

	return c.JSON(map[string]string{"status": "ok"})
}
