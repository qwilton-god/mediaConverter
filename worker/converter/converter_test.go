package converter

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap/zaptest"
)

func createTestImage(t *testing.T, width, height int, path string) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
}

func TestConverter_Convert_SimpleResize(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	createTestImage(t, 800, 600, inputPath)

	targetWidth := 400
	targetHeight := 300

	err := converter.Convert(inputPath, outputPath, "jpg", &targetWidth, &targetHeight, false)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode output image: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Expected dimensions 400x300, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestConverter_Convert_WithCrop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	createTestImage(t, 800, 600, inputPath)

	targetWidth := 300
	targetHeight := 300

	err := converter.Convert(inputPath, outputPath, "jpg", &targetWidth, &targetHeight, true)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode output image: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 300 || bounds.Dy() != 300 {
		t.Errorf("Expected dimensions 300x300, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestConverter_Convert_FormatConversion(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.png")

	createTestImage(t, 400, 300, inputPath)

	err := converter.Convert(inputPath, outputPath, "png", nil, nil, false)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	_, err = png.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode output as PNG: %v", err)
	}
}

func TestConverter_Convert_OnlyWidthSpecified(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	createTestImage(t, 800, 600, inputPath)

	targetWidth := 400

	err := converter.Convert(inputPath, outputPath, "jpg", &targetWidth, nil, false)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode output image: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 {
		t.Errorf("Expected width 400, got %d", bounds.Dx())
	}
}

func TestConverter_Convert_UnsupportedFormat(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.webp")

	createTestImage(t, 400, 300, inputPath)

	err := converter.Convert(inputPath, outputPath, "webp", nil, nil, false)
	if err == nil {
		t.Fatal("Expected error for unsupported format, got nil")
	}

	expectedErrMsg := "unsupported format: webp"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected '%s' error, got: %v", expectedErrMsg, err)
	}
}

func TestConverter_Convert_InvalidInputPath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.jpg")

	err := converter.Convert("/nonexistent/path.jpg", outputPath, "jpg", nil, nil, false)
	if err == nil {
		t.Fatal("Expected error for non-existent input file, got nil")
	}
}

func TestConverter_Convert_NoDimensionsPreservesOriginal(t *testing.T) {
	logger := zaptest.NewLogger(t)
	converter := NewConverter(logger)

	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.jpg")
	outputPath := filepath.Join(tmpDir, "output.jpg")

	createTestImage(t, 400, 300, inputPath)

	err := converter.Convert(inputPath, outputPath, "jpg", nil, nil, false)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode output image: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Expected dimensions 400x300 (original), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
