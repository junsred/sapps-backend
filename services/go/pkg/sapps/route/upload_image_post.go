package route

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"sapps/pkg/sapps/constant"
	maindb "sapps/pkg/sapps/lib/db/main"
	"sapps/pkg/sapps/middleware"

	"github.com/google/uuid"
	"go.uber.org/dig"
	_ "golang.org/x/image/webp"
)

type PostUploadImage struct {
	dig.In
	MainDB *maindb.MainDB
}

type PostUploadImageResponse struct {
	ImageID string `json:"image_id"`
}

func (r *PostUploadImage) Handler(c *middleware.RequestContext) error {
	file, err := c.FormFile("image")
	if err != nil {
		return c.Error(middleware.StatusBadRequest, "image file is required")
	}

	buffer, err := file.Open()
	if err != nil {
		return c.Error(middleware.StatusBadRequest, "invalid image")
	}
	defer buffer.Close()

	img, _, err := image.Decode(buffer)
	if err != nil {
		return c.Error(middleware.StatusBadRequest, "invalid image")
	}
	imageID := uuid.New().String()
	filename := imageID + ".jpg"
	out, err := os.Create(fmt.Sprintf("%s/cdn/img/%s", constant.WD_PATH, filename))
	if err != nil {
		return c.Error(middleware.StatusInternalServerError, "failed to save image")
	}
	defer out.Close()
	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 97})
	if err != nil {
		return c.Error(middleware.StatusInternalServerError, "failed to save image")
	}

	_, err = r.MainDB.Exec(c.Context(), "insert into images (id, user_id, size, content_type) values ($1, $2, $3, $4)",
		imageID, c.UserID(), file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		c.LogErr(err)
		return c.Error(middleware.StatusInternalServerError, "failed to save image info")
	}

	return c.JSON(PostUploadImageResponse{
		ImageID: imageID,
	})
}
