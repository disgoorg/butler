package butler

import (
	"sync"
	"time"

	"github.com/DisgoOrg/disgo/core/events"
)

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

type Component struct {
	Handler func(b *Butler, data []string, e *events.ComponentInteractionEvent)
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

func (c *Components) Get(name string) func(b *Butler, data []string, e *events.ComponentInteractionEvent) {
	c.RLock()
	defer c.RUnlock()
	component, ok := c.components[name]
	if !ok {
		return nil
	}
	return component.Handler
}

func (c *Components) Add(action string, handler func(b *Butler, data []string, e *events.ComponentInteractionEvent), timeout time.Duration) {
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
