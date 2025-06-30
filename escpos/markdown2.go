package escpos

import (
	"bytes"
	"fmt"
	"math"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

const (
	CharacterWidth uint8 = 48
)

type escr struct {
}

func (r *escr) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	// reg.Register(ast.KindBlockquote, r.renderBlockquote)
	// reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	// reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	// reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	// reg.Register(ast.KindList, r.renderList)
	// reg.Register(ast.KindListItem, r.renderListItem)
	// reg.Register(ast.KindParagraph, r.renderParagraph)
	// reg.Register(ast.KindTextBlock, r.renderTextBlock)
	// reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
	reg.Register(extast.KindTable, r.renderTable)
	reg.Register(extast.KindTableHeader, r.renderTableHeader)
	reg.Register(extast.KindTableRow, r.renderTableRow)
	reg.Register(extast.KindTableCell, r.renderTableCell)
	// inlines
	// reg.Register(ast.KindAutoLink, r.renderAutoLink)
	// reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	// reg.Register(ast.KindEmphasis, r.renderEmphasis)
	// reg.Register(ast.KindImage, r.renderImage)
	// reg.Register(ast.KindLink, r.renderLink)
	// reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
}

func (r *escr) renderDocument(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		writer.Write([]byte{gs, 'V', 'A', 0x00})
		return ast.WalkContinue, nil
	}
	return ast.WalkContinue, nil
}

func (r *escr) renderText(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		writer.Write([]byte{'\n'})
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment

	if n.IsRaw() {
		writer.Write(segment.Value(source))
	} else {
		value := segment.Value(source)
		writer.Write(value)
	}

	return ast.WalkContinue, nil
}

func (r *escr) renderString(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	writer.Write(n.Value)

	return ast.WalkContinue, nil
}

func (r *escr) renderHeading(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	level := uint8(6 - n.Level)

	if entering {
		writer.Write([]byte{gs, '!', (level << 4) | (level)})
	} else {
		writer.Write([]byte{gs, '!', (0 << 4) | (0)})
		writer.Write([]byte{0x0A})
	}

	return ast.WalkContinue, nil
}

func (r *escr) renderTable(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
func (r *escr) renderTableHeader(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		writer.Write([]byte{'\n'})
		return ast.WalkContinue, nil
	}
	return ast.WalkContinue, nil
}
func (r *escr) renderTableRow(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		writer.Write([]byte{'\n'})
		return ast.WalkContinue, nil
	}

	return ast.WalkContinue, nil
}

// Skips children as it only processes text
func (r *escr) renderTableCell(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	if node.ChildCount() != 1 {
		fmt.Println("Invalid table cell. Too many children")
		fmt.Println(node.ChildCount())
		return ast.WalkSkipChildren, nil
	}
	n := node.FirstChild().(*ast.Text)
	segment := n.Segment

	// Write the cell content
	var cellContent []byte
	if n.IsRaw() {
		cellContent = segment.Value(source)
	} else {
		cellContent = segment.Value(source)
	}
	writer.Write(cellContent)

	// Calculate column widths and current column index
	tableNode := node.Parent().Parent() // Get the table node
	var columnWidths []int
	var currentCol int

	// First pass: calculate maximum width for each column
	ast.Walk(tableNode, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == extast.KindTableRow {
			colIndex := 0
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				if child.Kind() == extast.KindTableCell {
					if child.FirstChild() != nil && child.FirstChild().Kind() == ast.KindText {
						cellText := child.FirstChild().(*ast.Text)
						cellLen := len(cellText.Segment.Value(source))

						// Extend columnWidths slice if necessary
						for len(columnWidths) <= colIndex {
							columnWidths = append(columnWidths, 0)
						}

						if cellLen > columnWidths[colIndex] {
							columnWidths[colIndex] = cellLen
						}
					}
					colIndex++
				}
			}
		}
		return ast.WalkContinue, nil
	})

	// Find current column index
	currentCol = 0
	for child := node.Parent().FirstChild(); child != nil && child != node; child = child.NextSibling() {
		if child.Kind() == extast.KindTableCell {
			currentCol++
		}
	}

	// Calculate tabs needed for this cell
	if currentCol < len(columnWidths) {
		currentCellLen := len(cellContent)
		maxColWidth := columnWidths[currentCol]

		// Calculate which tab stop this column should end at
		// Tab stops are at positions 8, 16, 24, 32, etc.
		tabStop := int(math.Ceil(float64(maxColWidth+1)/8.0)) * 8

		// Calculate how many characters we need to reach that tab stop
		charsToNextTab := tabStop - currentCellLen

		// Convert to number of tabs needed
		tabsNeeded := int(math.Ceil(float64(charsToNextTab) / 8.0))

		// Ensure at least 1 tab for column separation
		if tabsNeeded < 1 {
			tabsNeeded = 1
		}

		// Write the tabs
		for i := 0; i < tabsNeeded; i++ {
			writer.Write([]byte{0x09})
		}
	}

	return ast.WalkSkipChildren, nil
}

var EscposNodeRenderer renderer.NodeRenderer = &escr{}

func (e *Escpos) WriteMarkdown2(markdown []byte) (int, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithRenderer(
			renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(EscposNodeRenderer, 1))),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(markdown, &buf); err != nil {
		panic(err)
	}

	_, err := e.WriteRaw(buf.Bytes())
	if err != nil {
		return 0, err
	}
	return 0, nil
}
