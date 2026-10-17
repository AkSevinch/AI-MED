package docparse

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Format описывает поддерживаемый тип загрузки
type Format struct {
	Ext  string
	MIME string
	Text bool // true, если файл нужно привести к тексту перед анализом
}

var formats = map[string]Format{
	".pdf":  {Ext: ".pdf", MIME: "application/pdf"},
	".jpg":  {Ext: ".jpg", MIME: "image/jpeg"},
	".jpeg": {Ext: ".jpeg", MIME: "image/jpeg"},
	".png":  {Ext: ".png", MIME: "image/png"},
	".docx": {Ext: ".docx", MIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Text: true},
}

// Detect возвращает формат файла, ok=false если не поддерживается
func Detect(filename string) (Format, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	f, ok := formats[ext]
	return f, ok
}

// SupportedExtensions — список допустимых расширений для сообщений об ошибке
func SupportedExtensions() []string {
	return []string{".pdf", ".jpg", ".jpeg", ".png", ".docx"}
}

// DocxText извлекает текст из .docx на диске
func DocxText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read docx: %w", err)
	}
	return DocxTextBytes(b)
}

// DocxTextBytes извлекает текст из байтов .docx
func DocxTextBytes(b []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}

	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return "", fmt.Errorf("docx: word/document.xml not found")
	}

	rc, err := doc.Open()
	if err != nil {
		return "", fmt.Errorf("open document.xml: %w", err)
	}
	defer rc.Close()

	return extractWordText(rc)
}

func extractWordText(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	var sb strings.Builder
	var inText bool

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "p":
				if sb.Len() > 0 {
					sb.WriteByte('\n')
				}
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				sb.Write([]byte(t))
			}
		}
	}
	return strings.TrimSpace(sb.String()), nil
}
