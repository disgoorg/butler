package components

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/events"
)

var DocsActionComponent = butler.Component{
	Action:  "docs_action",
	Handler: handleDocsAction,
}

func handleDocsAction(b *butler.Butler, data []string, e *events.ComponentInteractionEvent) error {
	return nil
}
