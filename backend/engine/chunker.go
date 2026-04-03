package engine

import (
	"strings"
	"unicode"
)

type Chunker struct {
	chunkSize int
}

func NewChunker(chunkSize int) *Chunker {
	if chunkSize <= 0 {
		chunkSize = 3500
	}
	return &Chunker{chunkSize: chunkSize}
}

func (c *Chunker) Split(text string) []string {
	if len(text) <= c.chunkSize {
		return []string{text}
	}

	var chunks []string
	paragraphs := strings.Split(text, "\n\n")
	var current strings.Builder

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		if current.Len()+len(para)+2 > c.chunkSize {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}

			if len(para) > c.chunkSize {
				chunks = append(chunks, c.splitLongParagraph(para)...)
			} else {
				current.WriteString(para)
			}
		} else {
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(para)
		}
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	return chunks
}

func (c *Chunker) splitLongParagraph(para string) []string {
	var chunks []string

	sentenceEnds := []rune{'。', '！', '？', '.', '!', '?'}
	var sentences []string
	var current strings.Builder

	for _, r := range para {
		current.WriteRune(r)
		if containsRune(sentenceEnds, r) {
			sentences = append(sentences, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		sentences = append(sentences, current.String())
	}

	var chunk strings.Builder
	for _, s := range sentences {
		if chunk.Len()+len(s) > c.chunkSize && chunk.Len() > 0 {
			chunks = append(chunks, chunk.String())
			chunk.Reset()
		}
		chunk.WriteString(s)
	}
	if chunk.Len() > 0 {
		chunks = append(chunks, chunk.String())
	}

	return chunks
}

func containsRune(runes []rune, r rune) bool {
	for _, rr := range runes {
		if rr == r {
			return true
		}
	}
	return false
}

func IsCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x3040 && r <= 0x309F) ||
		(r >= 0x30A0 && r <= 0x30FF)
}

func IsChineseText(text string) bool {
	cjkCount := 0
	totalCount := 0
	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			continue
		}
		totalCount++
		if IsCJK(r) {
			cjkCount++
		}
	}
	if totalCount == 0 {
		return false
	}
	return float64(cjkCount)/float64(totalCount) > 0.3
}
