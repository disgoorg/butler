package butler

import (
	"strings"
	"sync"
	"time"

	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/log"
)

func (b *Butler) OnComponentInteraction(e *events.ComponentInteractionEvent) {
	data := strings.Split(e.Data.ID().String(), ":")
	action := data[0]
	if len(data) > 1 {
		data = append(data[:0], data[1:]...)
	}
	if handler := b.Components.Get(action); handler != nil {
		if err := handler(b, data, e); err != nil {
			b.Bot.Logger.Error("Error handling component: ", err)
		}
		return
	}
	log.Warnf("No handler for component with CustomID %s found", e.Data.ID())
}

func NewComponents() *Components {
	components := &Components{
		components: map[string]Component{},
	}
	components.startCleanup()
	return components
}

type Components struct {
	sync.RWMutex
	components map[string]Component
}

type ComponentHandlerFunc func(b *Butler, data []string, e *events.ComponentInteractionEvent) error

type Component struct {
	Handler ComponentHandlerFunc
	Timeout time.Time
}

func (c *Components) startCleanup() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			c.RLock()
			for action, component := range c.components {
				if component.Timeout.IsZero() {
					continue
				}
				if time.Now().After(component.Timeout) {
					c.Remove(action)
				}
			}
			c.RUnlock()
		}
	}()
}

func (c *Components) Get(name string) ComponentHandlerFunc {
	c.RLock()
	defer c.RUnlock()
	component, ok := c.components[name]
	if !ok {
		return nil
	}
	return component.Handler
}

func (c *Components) Add(action string, handler ComponentHandlerFunc, timeout time.Duration) {
	c.Lock()
	defer c.Unlock()

	t := time.Time{}
	if timeout > 0 {
		t = time.Now().Add(timeout)
	}
	c.components[action] = Component{
		Handler: handler,
		Timeout: t,
	}
}

func (c *Components) Remove(action string) {
	c.Lock()
	defer c.Unlock()
	delete(c.components, action)
}
