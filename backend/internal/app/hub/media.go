package hub

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/potibm/crewapparat/internal/app/config"
)

const (
	maxAvatarBytes = 5 * 1024 * 1024
	avatarSize     = 512
)

var (
	// ErrAvatarTooLarge is returned when the uploaded avatar exceeds maxAvatarBytes.
	ErrAvatarTooLarge = errors.New("avatar file too large")
	// ErrAvatarUnsupported is returned when the uploaded file is not a supported image.
	ErrAvatarUnsupported = errors.New("avatar file is not a supported image")
)

// respondWithAvatarStorageProblem maps avatar upload failures to HTTP responses:
// user-caused problems become a 400 with a helpful message, everything else a
// generic 500 so internal details stay out of the response.
func respondWithAvatarStorageProblem(c *gin.Context, err error) {
	if errors.Is(err, ErrAvatarTooLarge) || errors.Is(err, ErrAvatarUnsupported) {
		respondWithBadRequestProblem(c, err.Error())

		return
	}

	respondWithInternalServerProblem(c, "Failed to store avatar: "+err.Error())
}

//nolint:mnd // image processing constants
func saveAvatarFile(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > maxAvatarBytes {
		return "", fmt.Errorf("%w (max %d bytes)", ErrAvatarTooLarge, maxAvatarBytes)
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read uploaded file: %w", err)
	}

	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("%w: %s", ErrAvatarUnsupported, contentType)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrAvatarUnsupported, err)
	}

	out := resizeToSquareJPEG(img, avatarSize)

	mediaDir := filepath.Join(config.DataDirname, "media")
	if err := os.MkdirAll(mediaDir, config.DataDirPerm); err != nil {
		return "", fmt.Errorf("failed to create media directory: %w", err)
	}

	filename := uuid.NewString() + ".jpg"
	dest := filepath.Join(mediaDir, filename)

	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("failed to create media file: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, out, &jpeg.Options{Quality: 85}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	slog.Info("saveAvatarFile: stored avatar", "url", "/media/"+filename)

	return "/media/" + filename, nil
}

//nolint:mnd // image processing constants
func resizeToSquareJPEG(src image.Image, size int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	side := w
	if h < side {
		side = h
	}

	x0 := bounds.Min.X + (w-side)/2
	y0 := bounds.Min.Y + (h-side)/2
	cropped := image.NewRGBA(image.Rect(0, 0, side, side))
	draw.Draw(cropped, cropped.Bounds(), src, image.Point{x0, y0}, draw.Src)

	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	bilinearScale(dst, cropped, size)

	return dst
}

//nolint:mnd // image processing constants
func bilinearScale(dst, src *image.RGBA, size int) {
	srcBounds := src.Bounds()
	srcW, srcH := srcBounds.Dx(), srcBounds.Dy()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			sx := (float64(x)+0.5)*float64(srcW)/float64(size) - 0.5
			sy := (float64(y)+0.5)*float64(srcH)/float64(size) - 0.5
			x0 := int(math.Floor(sx))
			y0 := int(math.Floor(sy))

			if x0 < 0 {
				x0 = 0
			}

			if y0 < 0 {
				y0 = 0
			}

			x1 := x0 + 1
			if x1 >= srcW {
				x1 = srcW - 1
			}

			y1 := y0 + 1
			if y1 >= srcH {
				y1 = srcH - 1
			}

			fx := sx - float64(x0)
			fy := sy - float64(y0)

			if fx < 0 {
				fx = 0
			}

			if fy < 0 {
				fy = 0
			}

			c00 := src.RGBAAt(srcBounds.Min.X+x0, srcBounds.Min.Y+y0)
			c10 := src.RGBAAt(srcBounds.Min.X+x1, srcBounds.Min.Y+y0)
			c01 := src.RGBAAt(srcBounds.Min.X+x0, srcBounds.Min.Y+y1)
			c11 := src.RGBAAt(srcBounds.Min.X+x1, srcBounds.Min.Y+y1)

			r := bilinearChannel(c00.R, c10.R, c01.R, c11.R, fx, fy)
			g := bilinearChannel(c00.G, c10.G, c01.G, c11.G, fx, fy)
			b := bilinearChannel(c00.B, c10.B, c01.B, c11.B, fx, fy)
			a := bilinearChannel(c00.A, c10.A, c01.A, c11.A, fx, fy)
			dst.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

func bilinearChannel(c00, c10, c01, c11 uint8, fx, fy float64) uint8 {
	top := float64(c00)*(1-fx) + float64(c10)*fx
	bot := float64(c01)*(1-fx) + float64(c11)*fx
	val := top*(1-fy) + bot*fy

	return uint8(math.Round(val))
}
