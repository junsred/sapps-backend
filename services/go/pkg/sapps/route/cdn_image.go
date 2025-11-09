package route

import (
	"fmt"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"
	"os"

	"go.uber.org/dig"
)

type GetCDNImage struct {
	dig.In
	MainDB *maindb.MainDB
}

func (r *GetCDNImage) Handler(c *middleware.RequestContext) error {
	id := c.Params("id")
	image, err := os.Open(fmt.Sprintf("%s/cdn/img/%s", constant.WD_PATH, id))
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusNotFound)
	}
	return c.SendStream(image)
}
