package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmaupin/go-epub"
	"novella/backend/models"
)

type Exporter struct {
	format string
}

func NewExporter(format string) *Exporter {
	return &Exporter{format: format}
}

func (e *Exporter) Export(content string, title string, outputPath string, meta *models.EPUBMetadata) (string, error) {
	switch e.format {
	case "txt":
		return e.exportTXT(content, outputPath)
	case "md":
		return e.exportMD(content, title, outputPath)
	case "epub":
		return e.exportEPUB(content, title, outputPath, meta)
	default:
		return e.exportTXT(content, outputPath)
	}
}

func (e *Exporter) exportTXT(content string, outputPath string) (string, error) {
	if !strings.HasSuffix(outputPath, ".txt") {
		outputPath += ".txt"
	}
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (e *Exporter) exportMD(content string, title string, outputPath string) (string, error) {
	if !strings.HasSuffix(outputPath, ".md") {
		outputPath += ".md"
	}

	var md strings.Builder
	md.WriteString("# " + title + "\n\n")
	md.WriteString(content)

	if err := os.WriteFile(outputPath, []byte(md.String()), 0644); err != nil {
		return "", err
	}
	return outputPath, nil
}

func (e *Exporter) exportEPUB(content string, title string, outputPath string, meta *models.EPUBMetadata) (string, error) {
	if !strings.HasSuffix(outputPath, ".epub") {
		outputPath += ".epub"
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	book := epub.NewEpub(title)

	author := "Unknown"
	if meta != nil && meta.Author != "" {
		author = meta.Author
	}
	book.SetAuthor(author)

	if meta != nil && meta.Language != "" {
		book.SetLang(meta.Language)
	} else {
		book.SetLang("vi")
	}

	if meta != nil && meta.Description != "" {
		book.SetDescription(meta.Description)
	}

	if meta != nil && meta.CoverPath != "" {
		if _, err := os.Stat(meta.CoverPath); err == nil {
			coverExt := strings.ToLower(filepath.Ext(meta.CoverPath))

			coverRef, err := book.AddImage(meta.CoverPath, "cover"+coverExt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to add cover image: %v\n", err)
			} else {
				book.SetCover(coverRef, "")
			}
		} else {
			fmt.Fprintf(os.Stderr, "warning: cover image not found: %s\n", meta.CoverPath)
		}
	}

	chapters := strings.Split(content, "\n\n")
	var epubContent strings.Builder

	for _, chapter := range chapters {
		chapter = strings.TrimSpace(chapter)
		if chapter == "" {
			continue
		}

		epubContent.WriteString("<p>")
		epubContent.WriteString(chapter)
		epubContent.WriteString("</p>\n\n")
	}

	_, err := book.AddSection(epubContent.String(), title, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to add section: %w", err)
	}

	if err := book.Write(outputPath); err != nil {
		return "", fmt.Errorf("failed to write EPUB: %w", err)
	}

	return outputPath, nil
}
