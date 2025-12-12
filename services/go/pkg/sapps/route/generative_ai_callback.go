package route

import (
	"encoding/json"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"

	"go.uber.org/dig"
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
				resultURL = resultJSON.ResultURLs[0]
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
