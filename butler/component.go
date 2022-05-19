package butler

import (
	"strings"
	"time"

	"github.com/disgoorg/disgo/events"
)

func (b *Butler) SetupComponents(components ...Component) {
	for _, component := range components {
		b.Components[component.Action] = component
	}
}

func (b *Butler) OnComponentInteraction(e *events.ComponentInteractionEvent) {
	data := strings.Split(e.Data.CustomID().String(), ":")
	action := data[0]
	if len(data) > 1 {
		data = append(data[:0], data[1:]...)
	}
	if component, ok := b.Components[action]; ok {
		if err := component.Handler(b, data, e); err != nil {
			b.Client.Logger().Error("Error handling component: ", err)
		}
		return
	}
	b.Logger.Warnf("No handler for component with CustomID %s found", e.Data.CustomID())
}

type (
	ComponentHandlerFunc func(b *Butler, data []string, e *events.ComponentInteractionEvent) error
	Component            struct {
		Action  string
		Handler ComponentHandlerFunc
		Timeout time.Time
	}
)
