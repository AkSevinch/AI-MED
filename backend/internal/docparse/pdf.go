package docparse

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// PDFImage — растеризованная страница PDF
type PDFImage struct {
	Data []byte
	MIME string
}

// maxPDFPages — сколько первых страниц растеризуем (защита от гигантских файлов)
const maxPDFPages = 5

// RasterizePDF конвертирует первые страницы PDF в PNG через poppler (pdftoppm).
// Нужно потому, что OpenAI-совместимые провайдеры не принимают PDF как изображение.
func RasterizePDF(pdf []byte) ([]PDFImage, error) {
	dir, err := os.MkdirTemp("", "aimed-pdf-*")
	if err != nil {
		return nil, fmt.Errorf("pdf: temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	in := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(in, pdf, 0o600); err != nil {
		return nil, fmt.Errorf("pdf: write: %w", err)
	}

	prefix := filepath.Join(dir, "page")
	cmd := exec.Command("pdftoppm", "-png", "-r", "200",
		"-f", "1", "-l", fmt.Sprint(maxPDFPages), in, prefix)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pdf: pdftoppm: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	matches, err := filepath.Glob(prefix + "-*.png")
	if err != nil {
		return nil, err
	}
	// pdftoppm дополняет номер страницы нулями до одной ширины, поэтому
	// лексикографической сортировки достаточно для страниц 1..maxPDFPages
	sort.Strings(matches)
	if len(matches) == 0 {
		return nil, fmt.Errorf("pdf: не удалось растеризовать страницы")
	}

	out := make([]PDFImage, 0, len(matches))
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			return nil, err
		}
		out = append(out, PDFImage{Data: b, MIME: "image/png"})
	}
	return out, nil
}
