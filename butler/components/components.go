package components

import (
	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/events"
)

var Components = map[string]func(b *butler.Butler, data []string, e *events.ComponentInteractionEvent){
	"expand": handleExpand,
}
