// package main is a sample
package main

import (
	"fmt"
	"strings"
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

type operatorSet map[string]struct{}

var (
	_         Operator = (*Calc)(nil)
	operators          = operatorSet{
		"eq":  struct{}{},
		"in":  struct{}{},
		"con": struct{}{},
		"gt":  struct{}{},
		"lt":  struct{}{},
	}
)

// aad adds an op to the set
func (s operatorSet) add(op string) {
	s[op] = struct{}{}
}

// remove removes an op from the set
func (s operatorSet) remove(op string) {
	delete(s, op)
}

// has returns a boolean value describing if the op exists in the set
func (s operatorSet) has(op string) bool {
	op = strings.TrimSpace(strings.ToLower(op))
	_, ok := s[op]
	return ok
}

func main() {
	var wg sync.WaitGroup
	fmt.Println("sequential processing")
	o := &Calc{}
	o.Init(1, 2, 3, 4)
	o.Multiply(2)
	o.Add(1)
	o.Multiply(2)
	o.Display()

	o = &Calc{}
	o.Init(1, 2, 3, 4)
	wg.Add(3)
	fmt.Println("concurrent processing")
	go func() {
		defer wg.Done()
		o.Multiply(2)
	}()
	go func() {
		defer wg.Done()
		o.Add(1)
	}()
	go func() {
		defer wg.Done()
		o.Multiply(2)
	}()
	wg.Wait()
	o.Display()

	if operators.has("EQ") {
		fmt.Printf("found operator [%s] into the set of operators\n", "EQ")
	}
}
