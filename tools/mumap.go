package tools

import (
	"sync"

	"golang.org/x/exp/maps"
)

const (
	StatusInProgress string = "В процессе"
	StatusCompleted  string = "Завершено"
	StatusError      string = "Ошибка"
)

type Parsings struct {
	mx sync.RWMutex
	m  map[string]ParsData
}

type ParsData struct {
	Counter int
	Status  string
}

func NewParsings() *Parsings {
	return &Parsings{m: make(map[string]ParsData)}
}

func (c *Parsings) Get(key string) (ParsData, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	val, ok := c.m[key]
	return val, ok
}

func (c *Parsings) SetCounter(key string, value int) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if e, ok := c.m[key]; ok {
		e.Counter = value
		c.m[key] = e
	} else {
		c.m[key] = ParsData{Counter: value, Status: StatusInProgress}
	}
}

func (c *Parsings) Keys() []string {
	c.mx.Lock()
	defer c.mx.Unlock()
	keys := maps.Keys(c.m)

	return keys
}

func (c *Parsings) ChangeStatus(key string, status string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	if e, ok := c.m[key]; ok {
		e.Status = status
		c.m[key] = e
	}
}
