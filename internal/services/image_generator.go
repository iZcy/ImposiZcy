package services

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/sirupsen/logrus"
)

type ImageGeneratorService struct {
	logger *logrus.Logger
}

func NewImageGeneratorService(logger *logrus.Logger) *ImageGeneratorService {
	return &ImageGeneratorService{logger: logger}
}

func (s *ImageGeneratorService) GenerateFromHTML(ctx context.Context, html string, opts *models.RenderOptions) ([]byte, error) {
	width := 800
	height := 600
	quality := 90
	format := "png"
	scale := 2.0

	if opts != nil {
		if opts.Width > 0 {
			width = int(opts.Width)
		}
		if opts.Height > 0 {
			height = int(opts.Height)
		}
		if opts.Quality > 0 && opts.Quality <= 100 {
			quality = opts.Quality
		}
		if opts.Format != "" {
			format = opts.Format
		}
		if opts.Scale > 0 {
			scale = opts.Scale
		}
	}

	allocCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	chromeCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	timeoutCtx, cancel := context.WithTimeout(chromeCtx, 30*time.Second)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var imageBin []byte
			var err error

			if format == "jpeg" {
				imageBin, err = page.CaptureScreenshot().
					WithQuality(float64(quality) / 100.0).
					WithClip(&page.Viewport{
						X:      0,
						Y:      0,
						Width:  float64(width),
						Height: float64(height),
						Scale:  scale,
					}).Do(ctx)
			} else {
				imageBin, err = page.CaptureScreenshot().
					WithClip(&page.Viewport{
						X:      0,
						Y:      0,
						Width:  float64(width),
						Height: float64(height),
						Scale:  scale,
					}).Do(ctx)
			}

			if err != nil {
				return fmt.Errorf("failed to capture screenshot: %w", err)
			}
			buf = imageBin
			return nil
		}),
	); err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"format": format,
		"width":  width,
		"height": height,
		"scale":  scale,
		"size":   len(buf),
	}).Info("Image generated successfully")

	return buf, nil
}
