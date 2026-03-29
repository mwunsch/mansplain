// Package convert implements a ronn-format markdown to mdoc(7) converter
// using goldmark as the markdown parser.
package convert

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// ronnAngleRe matches ronn-format <variable> arguments (single word in angle brackets).
var ronnAngleRe = regexp.MustCompile(`<(\w+)>`)

// Convert transforms ronn-format markdown into mdoc(7) source.
func Convert(input []byte) ([]byte, error) {
	// Pre-process: convert ronn's <variable> notation to a placeholder
	// that goldmark won't treat as HTML. We'll convert back in rendering.
	processed := ronnAngleRe.ReplaceAll(input, []byte(`\fI$1\fR`))

	r := &mdocRenderer{}
	md := goldmark.New(
		goldmark.WithRenderer(
			renderer.NewRenderer(
				renderer.WithNodeRenderers(
					util.Prioritized(r, 1000),
				),
			),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(processed, &buf); err != nil {
		return nil, fmt.Errorf("converting markdown: %w", err)
	}

	return buf.Bytes(), nil
}

// mdocRenderer renders goldmark AST nodes as mdoc(7) macros.
type mdocRenderer struct {
	name        string
	section     int
	description string
	wroteHeader bool
	inList      bool
	listDepth   int
	needPp      bool // need .Pp before next paragraph
}

var titleRe = regexp.MustCompile(`^(\w[\w.-]*)\((\d)\)\s*(?:--|[-–—])\s*(.+)$`)
var xrefRe = regexp.MustCompile(`(\w[\w.-]*)\((\d)\)`)

func (r *mdocRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)

	// inlines
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
}

func (r *mdocRenderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)

	if !entering {
		return ast.WalkContinue, nil
	}

	text := extractText(node, source)

	if n.Level == 1 && !r.wroteHeader {
		r.wroteHeader = true
		if m := titleRe.FindStringSubmatch(text); m != nil {
			r.name = m[1]
			r.section = int(m[2][0] - '0')
			r.description = m[3]
		} else {
			// Plain H1 — use as name, default section 1
			r.name = strings.TrimSpace(text)
			r.section = 1
		}
		upper := strings.ToUpper(r.name)
		fmt.Fprintf(w, ".Dd $Mdocdate$\n.Dt %s %d\n.Os\n.Sh NAME\n.Nm %s\n", upper, r.section, r.name)
		if r.description != "" {
			fmt.Fprintf(w, ".Nd %s\n", r.description)
		}
		return ast.WalkSkipChildren, nil
	}

	if n.Level == 2 {
		fmt.Fprintf(w, ".Sh %s\n", strings.ToUpper(text))
		r.needPp = false
		return ast.WalkSkipChildren, nil
	}

	// H3+
	fmt.Fprintf(w, ".Ss %s\n", text)
	r.needPp = false
	return ast.WalkSkipChildren, nil
}

func (r *mdocRenderer) renderParagraph(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		if r.needPp && !r.inList {
			w.WriteString(".Pp\n")
		}
		// Don't emit .Pp inside list items — it creates unwanted blank lines
	} else {
		w.WriteByte('\n')
		r.needPp = true
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(".Bd -literal -offset indent\n")
		r.writeLines(w, source, node)
		w.WriteString(".Ed\n")
		r.needPp = true
	}
	return ast.WalkSkipChildren, nil
}

func (r *mdocRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return r.renderFencedCodeBlock(w, source, node, entering)
}

func (r *mdocRenderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	if entering {
		if isDefinitionList(node, source) {
			w.WriteString(".Bl -tag -width indent\n")
		} else if n.IsOrdered() {
			w.WriteString(".Bl -enum\n")
		} else {
			w.WriteString(".Bl -bullet\n")
		}
		r.inList = true
		r.listDepth++
	} else {
		w.WriteString(".El\n")
		r.listDepth--
		if r.listDepth == 0 {
			r.inList = false
		}
		r.needPp = true
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderListItem(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	// Check if this is a definition list item (first line ends with ':')
	text := extractFirstLineText(node, source)
	if defTerm, ok := parseDefinitionTerm(text); ok {
		fmt.Fprintf(w, ".It %s\n", convertInlineMarkup(defTerm))
		// Extract and write the description (everything after the ':')
		fullText := string(extractFullText(node.FirstChild(), source))
		if _, rest, found := strings.Cut(fullText, ":"); found {
			rest = strings.TrimSpace(rest)
			if rest != "" {
				fmt.Fprintf(w, "%s\n", rest)
			}
		}
		// Render any subsequent paragraphs (after the first) as children
		first := true
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if first {
				first = false
				continue
			}
			// For subsequent paragraphs in the list item, extract text
			if p, ok := child.(*ast.Paragraph); ok {
				fmt.Fprintf(w, "%s\n", strings.TrimSpace(string(extractFullText(p, source))))
			}
		}
		return ast.WalkSkipChildren, nil
	}

	w.WriteString(".It\n")
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(".Bd -offset indent\n")
	} else {
		w.WriteString(".Ed\n")
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderThematicBreak(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString(".Pp\n")
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderTextBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		w.WriteByte('\n')
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Skip HTML blocks — not relevant for mdoc
	return ast.WalkSkipChildren, nil
}

// Inline renderers

func (r *mdocRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	text := string(n.Segment.Value(source))

	// Convert man page references like name(1) to .Xr on their own line.
	// Handle "name(1)," → ".Xr name 1 ," (comma as separate mdoc argument).
	if xrefRe.MatchString(text) {
		xrefWithTrailing := regexp.MustCompile(`(\w[\w.-]*)\((\d)\)([,;]?)`)
		text = xrefWithTrailing.ReplaceAllStringFunc(text, func(match string) string {
			m := xrefWithTrailing.FindStringSubmatch(match)
			if m[3] != "" {
				return fmt.Sprintf("\n.Xr %s %s %s", m[1], m[2], m[3])
			}
			return fmt.Sprintf("\n.Xr %s %s", m[1], m[2])
		})
	}

	w.WriteString(text)
	if n.SoftLineBreak() {
		w.WriteByte('\n')
	}
	if n.HardLineBreak() {
		w.WriteString("\n.br\n")
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.String)
		w.Write(n.Value)
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderCodeSpan(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		text := extractText(node, source)
		w.WriteString(convertCodeSpan(text))
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	if n.Level == 2 {
		// **bold**
		if entering {
			w.WriteString("\\fB")
		} else {
			w.WriteString("\\fR")
		}
	} else {
		// _italic_
		if entering {
			w.WriteString("\\fI")
		} else {
			w.WriteString("\\fR")
		}
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		text := extractText(node, source)
		// Check if the link text is a man reference
		if m := xrefRe.FindStringSubmatch(text); m != nil && m[0] == text {
			fmt.Fprintf(w, ".Xr %s %s", m[1], m[2])
			return ast.WalkSkipChildren, nil
		}
		w.WriteString(text)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.AutoLink)
		label := string(n.Label(source))
		fmt.Fprintf(w, ".Lk %s", label)
		return ast.WalkSkipChildren, nil
	}
	return ast.WalkContinue, nil
}

func (r *mdocRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// Images don't translate to mdoc
	return ast.WalkSkipChildren, nil
}

func (r *mdocRenderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		n := node.(*ast.RawHTML)
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			tag := string(seg.Value(source))
			// Handle ronn's <var> convention for arguments
			tag = strings.TrimSpace(tag)
			if strings.HasPrefix(tag, "<") && !strings.HasPrefix(tag, "</") {
				// Extract content of <var>text</var> — but goldmark splits
				// these across multiple nodes. Handle the opening tag by
				// starting italic for argument display.
				inner := strings.TrimPrefix(tag, "<")
				inner = strings.TrimSuffix(inner, ">")
				if inner == "var" || inner == "em" {
					w.WriteString("\\fI")
				}
			} else if strings.HasPrefix(tag, "</") {
				inner := strings.TrimPrefix(tag, "</")
				inner = strings.TrimSuffix(inner, ">")
				if inner == "var" || inner == "em" {
					w.WriteString("\\fR")
				}
			}
		}
	}
	return ast.WalkContinue, nil
}

// Helpers

func (r *mdocRenderer) writeLines(w util.BufWriter, source []byte, node ast.Node) {
	l := node.Lines().Len()
	for i := 0; i < l; i++ {
		line := node.Lines().At(i)
		w.Write(line.Value(source))
	}
}

// extractText walks a node's children and extracts all text content.
func extractText(node ast.Node, source []byte) string {
	var buf strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if t, ok := child.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
		} else if cs, ok := child.(*ast.CodeSpan); ok {
			// Extract code span text
			for gc := cs.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if t, ok := gc.(*ast.Text); ok {
					buf.Write(t.Segment.Value(source))
				}
			}
		} else {
			buf.WriteString(extractText(child, source))
		}
	}
	return buf.String()
}

// extractFullText extracts all text from a node including soft line breaks.
func extractFullText(node ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			buf.Write(t.Segment.Value(source))
			if t.SoftLineBreak() {
				buf.WriteByte(' ')
			}
		}
		return ast.WalkContinue, nil
	})
	return buf.Bytes()
}

// extractFirstLineText gets the text content of a list item's first paragraph,
// preserving backtick markup for inline code spans.
func extractFirstLineText(node ast.Node, source []byte) string {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if p, ok := child.(*ast.Paragraph); ok {
			return extractMarkupText(p, source)
		}
		if tb, ok := child.(*ast.TextBlock); ok {
			return extractMarkupText(tb, source)
		}
	}
	return ""
}

// extractMarkupText preserves backtick and emphasis markup while extracting text.
func extractMarkupText(node ast.Node, source []byte) string {
	var buf strings.Builder
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			buf.Write(n.Segment.Value(source))
		case *ast.CodeSpan:
			buf.WriteByte('`')
			buf.WriteString(extractText(n, source))
			buf.WriteByte('`')
		case *ast.Emphasis:
			if n.Level == 2 {
				buf.WriteString("**")
				buf.WriteString(extractText(n, source))
				buf.WriteString("**")
			} else {
				buf.WriteByte('_')
				buf.WriteString(extractText(n, source))
				buf.WriteByte('_')
			}
		case *ast.RawHTML:
			// Reconstruct <var> tags
			for i := 0; i < n.Segments.Len(); i++ {
				seg := n.Segments.At(i)
				buf.Write(seg.Value(source))
			}
		default:
			buf.WriteString(extractText(child, source))
		}
	}
	return buf.String()
}

// isDefinitionList checks if any list item's first line ends with ':'.
func isDefinitionList(node ast.Node, source []byte) bool {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if li, ok := child.(*ast.ListItem); ok {
			text := extractFirstLineText(li, source)
			if _, ok := parseDefinitionTerm(text); ok {
				return true
			}
		}
	}
	return false
}

// parseDefinitionTerm checks if text matches a definition list term (ends with ':').
func parseDefinitionTerm(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if idx := strings.Index(text, ":"); idx > 0 {
		// Check that the colon terminates the "term" portion
		term := strings.TrimSpace(text[:idx])
		if term != "" {
			return term, true
		}
	}
	return "", false
}

// convertCodeSpan converts backtick-delimited text to appropriate mdoc markup.
func convertCodeSpan(text string) string {
	// Flags: starts with - or --
	if strings.HasPrefix(text, "--") {
		flag := strings.TrimPrefix(text, "--")
		if arg, found := splitFlagArg(flag); found {
			return fmt.Sprintf(".Fl -%s Ar %s", arg[0], arg[1])
		}
		return fmt.Sprintf("\\fB--%s\\fR", flag)
	}
	if strings.HasPrefix(text, "-") && len(text) == 2 {
		return fmt.Sprintf(".Fl %s", text[1:])
	}

	// Paths
	if strings.Contains(text, "/") || strings.HasPrefix(text, "~") || strings.HasPrefix(text, ".") {
		return fmt.Sprintf(".Pa %s", text)
	}

	// Environment variables (ALL_CAPS with underscores)
	if isEnvVar(text) {
		return fmt.Sprintf(".Ev %s", text)
	}

	// Default: bold
	return fmt.Sprintf("\\fB%s\\fR", text)
}

// convertInlineMarkup converts inline ronn markdown to mdoc for definition term lines.
// Uses macro names without leading dots since they appear as arguments to .It.
func convertInlineMarkup(text string) string {
	// Convert backtick spans
	result := regexp.MustCompile("`([^`]+)`").ReplaceAllStringFunc(text, func(m string) string {
		inner := m[1 : len(m)-1]
		return convertCodeSpanForMacroLine(inner)
	})
	// Convert <angle> to Ar
	result = regexp.MustCompile(`<(\w+)>`).ReplaceAllString(result, "Ar $1")
	return result
}

// convertCodeSpanForMacroLine converts backtick text to mdoc macro arguments
// (without leading dots, for use inside .It lines).
func convertCodeSpanForMacroLine(text string) string {
	if strings.HasPrefix(text, "--") {
		flag := strings.TrimPrefix(text, "--")
		return fmt.Sprintf("Fl -%s", flag)
	}
	if strings.HasPrefix(text, "-") && len(text) == 2 {
		return fmt.Sprintf("Fl %s", text[1:])
	}
	if isEnvVar(text) {
		return fmt.Sprintf("Ev %s", text)
	}
	return fmt.Sprintf("\\fB%s\\fR", text)
}

func splitFlagArg(flag string) ([2]string, bool) {
	for _, sep := range []string{"=", " "} {
		if i := strings.Index(flag, sep); i > 0 {
			return [2]string{flag[:i], flag[i+1:]}, true
		}
	}
	return [2]string{}, false
}

func isEnvVar(s string) bool {
	if len(s) < 2 {
		return false
	}
	for _, c := range s {
		if c != '_' && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}
