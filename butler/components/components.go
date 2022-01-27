package components

import (
	"github.com/DisgoOrg/disgo-butler/butler"
	"github.com/DisgoOrg/disgo/core/events"
)

var Components = map[string]func(b *butler.Butler, data []string, e *events.ComponentInteractionEvent){
	"expand": handleExpand,
}
