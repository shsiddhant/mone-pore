package markdown

import (
	"bytes"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

var (
	mdParser = parser.New(
		parser.WithExtensions(
			extension.GFMParser, // Github flavoured markdown
		),
		parser.WithAutoHeadingID(),
	)

	mdRenderer = html.New(
		html.WithExtensions(
			extension.GFMHTMLRenderer,
		),
		html.WithHardWraps(),
	)
)

// ToHTML converts a markdown string into a safe templ Component.
// If markdown renderer gives and error, the plain markdown text is rendered.
func ToHTML(mdString string) templ.Component {

	source := []byte(mdString)

	doc := mdParser.Parse(source)

	var buf bytes.Buffer

	if err := mdRenderer.Render(&buf, source, doc); err != nil {
		return templ.Raw(mdString)
	}

	return templ.Raw(buf.String())

}
