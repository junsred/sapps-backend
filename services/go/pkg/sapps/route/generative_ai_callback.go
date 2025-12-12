package route

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"time"

	"github.com/google/uuid"
	"go.uber.org/dig"
	_ "golang.org/x/image/webp"
)

type PostGenerativeAICallback struct {
	dig.In
	MainDB *maindb.MainDB
}

type KieCallbackRequest struct {
	Code int `json:"code"`
	Data struct {
		CompleteTime    int64  `json:"completeTime"`
		ConsumeCredits  int    `json:"consumeCredits"`
		CostTime        int    `json:"costTime"`
		CreateTime      int64  `json:"createTime"`
		Model           string `json:"model"`
		Param           string `json:"param"`
		RemainedCredits int    `json:"remainedCredits"`
		ResultJSON      string `json:"resultJson"`
		State           string `json:"state"`
		TaskID          string `json:"taskId"`
		UpdateTime      int64  `json:"updateTime"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type KieResultJSON struct {
	ResultURLs []string `json:"resultUrls"`
}

func downloadAndSaveImage(externalURL string) (string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(externalURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "kie-image-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save temp image: %w", err)
	}

	tempFile.Seek(0, 0)
	img, _, err := image.Decode(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	imageID := uuid.New().String()
	filename := imageID + ".jpg"
	outPath := fmt.Sprintf("%s/cdn/img/%s", constant.WD_PATH, filename)

	out, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 97})
	if err != nil {
		return "", fmt.Errorf("failed to encode jpeg: %w", err)
	}

	localURL := fmt.Sprintf("%s/cdn/img/%s.jpg", constant.API_URL, imageID)
	return localURL, nil
}

func (r *PostGenerativeAICallback) Handler(c *middleware.RequestContext) error {
	var req KieCallbackRequest
	if err := c.BodyParser(&req); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusBadRequest, "invalid request body")
	}

	if req.Data.TaskID == "" {
		return c.Error(middleware.StatusBadRequest, "task_id is required")
	}

	status := "failed"
	if req.Code == 200 && req.Data.State == "success" {
		status = "completed"
	}

	var resultURL string
	if req.Data.ResultJSON != "" {
		var resultJSON KieResultJSON
		if err := json.Unmarshal([]byte(req.Data.ResultJSON), &resultJSON); err == nil {
			if len(resultJSON.ResultURLs) > 0 {
				externalURL := resultJSON.ResultURLs[0]
				localURL, err := downloadAndSaveImage(externalURL)
				if err != nil {
					c.LogErr(err)
					status = "failed"
				} else {
					resultURL = localURL
				}
			}
		}
	}

	_, err := r.MainDB.Exec(c.Context(),
		`UPDATE generative_ai_tasks 
		 SET status = $1, result_url = $2, completed_at = NOW(), raw_response = $3
		 WHERE task_id = $4`,
		status, resultURL, req.Data.ResultJSON, req.Data.TaskID)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to update task")
	}

	return c.JSON(map[string]string{"status": "ok"})
}
