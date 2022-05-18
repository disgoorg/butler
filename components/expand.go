package components

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/events"
)

var ExpandComponent = butler.Component{
	Action:  "expand",
	Handler: handleExpand,
}

func handleExpand(b *butler.Butler, data []string, e *events.ComponentInteractionEvent) error {
	return nil
}
