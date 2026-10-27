package utils

import (
	"os"

	"github.com/pgavlin/goldmark"
	goldmark_renderer "github.com/pgavlin/goldmark/renderer"
	"github.com/pgavlin/goldmark/text"
	"github.com/pgavlin/goldmark/util"
	"github.com/pgavlin/markdown-kit/renderer"
	"github.com/pgavlin/markdown-kit/styles"
)

func PrintMd(source string) error {
	document := goldmark.DefaultParser().Parse(text.NewReader([]byte(source)))

	r := renderer.New(
		renderer.WithTheme(styles.AutoTheme()),
		renderer.WithWordWrap(80),
		renderer.WithSoftBreak(true),
		renderer.WithPad(true),
		renderer.WithHyperlinks(true),
		renderer.WithImages(false, 0, "."),
	)

	rr := goldmark_renderer.NewRenderer(
		goldmark_renderer.WithNodeRenderers(util.Prioritized(r, 100)),
	)

	if err := rr.Render(os.Stdout, []byte(source), document); err != nil {
		return err
	}
	return nil
}
