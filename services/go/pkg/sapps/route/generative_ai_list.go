package route

import (
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"

	"go.uber.org/dig"
)

type GetGenerativeAIList struct {
	dig.In
	MainDB *maindb.MainDB
}

type GenerativeAIListItem struct {
	ID        string  `json:"id"`
	ResultURL *string `json:"result_url,omitempty"`
}

type GetGenerativeAIListResponse struct {
	Generations []GenerativeAIListItem `json:"generations"`
}

func (r *GetGenerativeAIList) Handler(c *middleware.RequestContext) error {
	rows, err := r.MainDB.Query(c.Context(),
		`SELECT id, result_url 
		 FROM generative_ai_tasks 
		 WHERE user_id = $1 AND status = 'completed'
		 ORDER BY created_at DESC`,
		c.UserID())
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to fetch generations")
	}
	defer rows.Close()

	generations := []GenerativeAIListItem{}
	for rows.Next() {
		var item GenerativeAIListItem
		if err := rows.Scan(&item.ID, &item.ResultURL); err != nil {
			c.LogErr(err)
			continue
		}
		generations = append(generations, item)
	}

	return c.JSON(GetGenerativeAIListResponse{
		Generations: generations,
	})
}
