package converter

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	"go.uber.org/zap"
)

type Converter struct {
	logger *zap.Logger
}

func NewConverter(logger *zap.Logger) *Converter {
	return &Converter{logger: logger}
}

func (c *Converter) Convert(inputPath, outputPath, outputFormat string, targetWidth, targetHeight *int, crop bool) error {
	c.logger.Info("Starting conversion",
		zap.String("input", inputPath),
		zap.String("output", outputPath),
		zap.String("format", outputFormat),
	)

	src, err := imaging.Open(inputPath)
	if err != nil {
		c.logger.Error("Failed to open image",
			zap.String("path", inputPath),
			zap.Error(err),
		)
		return fmt.Errorf("failed to open image: %w", err)
	}

	var processedImage *image.NRGBA

	if targetWidth != nil || targetHeight != nil {
		width := targetWidth
		height := targetHeight

		if width == nil {
			w := src.Bounds().Dx()
			width = &w
		}
		if height == nil {
			h := src.Bounds().Dy()
			height = &h
		}

		c.logger.Info("Resizing image",
			zap.Int("width", *width),
			zap.Int("height", *height),
			zap.Bool("crop", crop),
		)

		if crop {
			processedImage = imaging.Fill(src, *width, *height, imaging.Center, imaging.Lanczos)
		} else {
			processedImage = imaging.Resize(src, *width, *height, imaging.Lanczos)
		}
	} else {
		processedImage = imaging.Clone(src)
	}

	if outputFormat != "" {
		switch outputFormat {
		case "jpg", "jpeg":
			if err := imaging.Save(processedImage, outputPath, imaging.JPEGQuality(85)); err != nil {
				c.logger.Error("Failed to save JPEG",
					zap.String("path", outputPath),
					zap.Error(err),
				)
				return fmt.Errorf("failed to save JPEG: %w", err)
			}
		case "png":
			if err := imaging.Save(processedImage, outputPath); err != nil {
				c.logger.Error("Failed to save PNG",
					zap.String("path", outputPath),
					zap.Error(err),
				)
				return fmt.Errorf("failed to save PNG: %w", err)
			}
		default:
			err := fmt.Errorf("unsupported format: %s", outputFormat)
			c.logger.Error("Unsupported format", zap.Error(err))
			return err
		}
	} else {
		if err := imaging.Save(processedImage, outputPath); err != nil {
			c.logger.Error("Failed to save image",
				zap.String("path", outputPath),
				zap.Error(err),
			)
			return fmt.Errorf("failed to save image: %w", err)
		}
	}

	c.logger.Info("Conversion completed",
		zap.String("output", outputPath),
	)

	return nil
}
