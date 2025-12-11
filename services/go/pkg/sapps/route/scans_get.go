package route

import (
	"fmt"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"time"

	"go.uber.org/dig"
)

type GetScans struct {
	dig.In
	MainDB *maindb.MainDB
}

type ScanItem struct {
	ScanID    string `json:"scan_id"`
	ImageURL  string `json:"image_url"`
	CreatedAt int64  `json:"created_at"`
}

type GetScansResponse struct {
	Scans []ScanItem `json:"scans"`
}

func (r *GetScans) Handler(c *middleware.RequestContext) error {
	rows, err := r.MainDB.Query(c.Context(),
		"SELECT scan_id, image_id, created_at FROM scans WHERE user_id = $1 ORDER BY created_at DESC",
		c.UserID())
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to fetch scans")
	}
	defer rows.Close()

	scans := []ScanItem{}
	for rows.Next() {
		var scanID string
		var imageID string
		var createdAt time.Time
		if err := rows.Scan(&scanID, &imageID, &createdAt); err != nil {
			c.LogErr(err)
			continue
		}
		imageURL := fmt.Sprintf("%s/cdn/img/%s.jpg", constant.API_URL, imageID)
		scans = append(scans, ScanItem{
			ScanID:    scanID,
			ImageURL:  imageURL,
			CreatedAt: createdAt.Unix(),
		})
	}

	return c.JSON(GetScansResponse{
		Scans: scans,
	})
}
