package escpos

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func (e *Escpos) ResetStyles() {
	e.Style = Style{
		Bold:       false,
		Width:      1,
		Height:     1,
		Reverse:    false,
		UpsideDown: false,
		Rotate:     false,
		Justify:    JustifyLeft,
	}
}

// WriteMarkdown parses markdown text and converts it to ESC/POS commands
func (e *Escpos) WriteMarkdown(markdown string) (int, error) {
	lines := strings.Split(markdown, "\n")
	totalWritten := 0

	// Save current style to restore later
	originalStyle := e.Style

	for i, line := range lines {
		written, err := e.processMarkdownLine(line)
		if err != nil {
			return totalWritten, err
		}
		totalWritten += written

		// Add line feed except for the last line
		if i < len(lines)-1 {
			lfWritten, err := e.LineFeed()
			if err != nil {
				return totalWritten, err
			}
			totalWritten += lfWritten
		}
	}
	// Restore original style
	e.Style = originalStyle
	return totalWritten, nil
}

// processMarkdownLine processes a single line of markdown
func (e *Escpos) processMarkdownLine(line string) (int, error) {
	e.ResetStyles()
	// Trim leading and trailing whitespace
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return 0, nil
	}

	// Handle headers
	if headerMatch := regexp.MustCompile(`^(#{1,6})\s+(.+)$`).FindStringSubmatch(line); headerMatch != nil {
		return e.processHeader(len(headerMatch[1]), headerMatch[2])
	}

	// Handle horizontal rules
	if matched, _ := regexp.MatchString(`^-{3,}$|^\*{3,}$|^_{3,}$`, line); matched {
		return e.processHorizontalRule()
	}

	// Handle unordered lists
	if listMatch := regexp.MustCompile(`^[\s]*[-\*]\s+(.+)$`).FindStringSubmatch(line); listMatch != nil {
		return e.processUnorderedListItem(listMatch[1])
	}

	// Handle ordered lists
	if listMatch := regexp.MustCompile(`^[\s]*(\d+)\.\s+(.+)$`).FindStringSubmatch(line); listMatch != nil {
		num, _ := strconv.Atoi(listMatch[1])
		return e.processOrderedListItem(num, listMatch[2])
	}

	// Handle code blocks (simplified - just detect ``` lines)
	if matched, _ := regexp.MatchString("^```", line); matched {
		// For simplicity, just print the line as-is for code block delimiters
		return e.Write(line)
	}

	// Handle regular text with inline formatting
	return e.processInlineFormatting(line)
}

// processHeader handles markdown headers
func (e *Escpos) processHeader(level int, text string) (int, error) {
	// Reset style
	e.ResetStyles()

	// Set size based on header level (larger for smaller level numbers)
	switch level {
	case 1:
		e.Size(5, 5) // Largest
		e.Bold(true)
	case 2:
		e.Size(4, 4)
		e.Bold(true)
	case 3:
		e.Size(3, 3)
		e.Bold(true)
	case 4:
		e.Size(2, 2)
		e.Bold(true)
	case 5:
		e.Size(2, 1)
		e.Bold(true)
	case 6:
		e.Size(1, 1)
		e.Bold(true)
		e.Underline(1)
	}

	// Center align headers
	// e.Justify(JustifyLeft)

	written, err := e.Write(text)
	if err != nil {
		return written, err
	}

	// Add extra line feed after headers
	lfWritten, err := e.LineFeed()
	if err != nil {
		return written, err
	}

	// Reset style and center align
	e.ResetStyles()

	return written + lfWritten, nil
}

// processHorizontalRule handles horizontal rules
func (e *Escpos) processHorizontalRule() (int, error) {
	// Reset style and center align
	e.ResetStyles()
	e.Justify(JustifyCenter)

	return e.Write("----------------------------------------")
}

// processUnorderedListItem handles unordered list items
func (e *Escpos) processUnorderedListItem(text string) (int, error) {
	// Reset style and left align
	e.ResetStyles()
	e.Justify(JustifyLeft)

	// Process inline formatting in the list item text
	formattedText, err := e.formatInlineText(text)
	if err != nil {
		return 0, err
	}

	return e.Write("• " + formattedText)
}

// processOrderedListItem handles ordered list items
func (e *Escpos) processOrderedListItem(num int, text string) (int, error) {
	// Reset style and left align
	e.ResetStyles()
	e.Justify(JustifyLeft)

	// Process inline formatting in the list item text
	formattedText, err := e.formatInlineText(text)
	if err != nil {
		return 0, err
	}

	return e.Write(fmt.Sprintf("%d. %s", num, formattedText))
}

// processInlineFormatting handles regular text with inline formatting
func (e *Escpos) processInlineFormatting(line string) (int, error) {
	// Reset style and left align
	e.ResetStyles()
	e.Justify(JustifyLeft)

	formattedText, err := e.formatInlineText(line)
	if err != nil {
		return 0, err
	}

	return e.Write(formattedText)
}

// formatInlineText processes inline markdown formatting and returns plain text
// Note: This is a simplified implementation that removes markdown syntax
// For a full implementation, you'd want to apply formatting in real-time
func (e *Escpos) formatInlineText(text string) (string, error) {
	// Remove bold formatting (**text** and __text__)
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	text = boldRegex.ReplaceAllString(text, "$1$2")

	// Remove italic formatting (*text* and _text_)
	italicRegex := regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	text = italicRegex.ReplaceAllString(text, "$1$2")

	// Remove strikethrough (~~text~~)
	strikeRegex := regexp.MustCompile(`~~([^~]+)~~`)
	text = strikeRegex.ReplaceAllString(text, "$1")

	// Remove inline code (`text`)
	codeRegex := regexp.MustCompile("`([^`]+)`")
	text = codeRegex.ReplaceAllString(text, "$1")

	return text, nil
}
