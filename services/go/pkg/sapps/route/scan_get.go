package route

import (
	"encoding/json"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"time"

	"go.uber.org/dig"
)

type GetScan struct {
	dig.In
	MainDB *maindb.MainDB
}

type GetScanResponse struct {
	ScanID    string   `json:"scan_id"`
	Data      ScanData `json:"data"`
	CreatedAt int64    `json:"created_at"`
}

func (r *GetScan) Handler(c *middleware.RequestContext) error {
	scanID := c.Params("id")
	if scanID == "" {
		return c.Error(middleware.StatusBadRequest, "scan id is required")
	}

	var dataBytes []byte
	var createdAt time.Time
	err := r.MainDB.QueryRow(c.Context(),
		"SELECT data, created_at FROM scans WHERE scan_id = $1 AND user_id = $2",
		scanID, c.UserID()).Scan(&dataBytes, &createdAt)
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusNotFound, "scan not found")
	}

	var data ScanData
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to parse scan data")
	}

	return c.JSON(GetScanResponse{
		ScanID:    scanID,
		Data:      data,
		CreatedAt: createdAt.Unix(),
	})
}
