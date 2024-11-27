// package main is a sample
package main

import (
	"fmt"
	"sync"
)

type Operator interface {
	Init(...int)
	Multiply(int)
	Add(int)
	Display()
}

type Calc struct {
	mu     sync.Mutex
	values []int
}

func (c *Calc) Init(values ...int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = values
}

func (c *Calc) Multiply(k int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < len(c.values); i++ {
		c.values[i] = k * c.values[i]
	}
}
func (c *Calc) Add(k int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < len(c.values); i++ {
		c.values[i] = k + c.values[i]
	}
}

func (c *Calc) Display() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.values {
		fmt.Println(v)
	}
}

var _ Operator = (*Calc)(nil)

func main() {
	o := &Calc{}
	o.Init(1, 2, 3, 4)
	o.Multiply(2)
	o.Add(1)
	o.Multiply(2)
	o.Display()
}
