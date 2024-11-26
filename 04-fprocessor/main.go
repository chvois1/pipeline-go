// Package main strongly supports the pipeline concurrency pattern.
// Pipelines are using channels.
// Thus at the end of a pipeline, we can use a range statement to extract the values.
// So at each stage, we can safely execute concurrently because our inputs and outputs are safe in concurrent contexts.
// Each stage of the pipeline is executing concurrently, means any stage only need wait for its inputs, and to be able to send its outputs.
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
)

const RxEvent = "FileReceived"

type Message struct {
	Id          int    `json:"id,omitempty"`
	State       int    `json:"state"`
	Description string `json:"description,omitempty"`
}

type Job struct {
	Id     string `json:"id,omitempty"`
	Cmd    string `json:"cmd,omitempty"`
	Status string `json:"status,omitempty"`
}

type Pipe struct {
	mu     sync.RWMutex
	subs   map[string][]chan Message
	closed bool
}

func NewPipe() *Pipe {
	log.Println("=> NewPipe")
	log.Println("<= NewPipe")
	return &Pipe{
		subs:   make(map[string][]chan Message),
		closed: false,
	}
}

// One of the benefits of using Go is having an application compiled into a single self-contained binary.
// Having a way to embed files in Go programs is the missing piece that helps us keep a single binary and bundle out static content.
//
//go:embed web/public
var frontend embed.FS
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Resolve cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// ensure the program exits cleanly and never lmeaks goroutines

	done := make(chan interface{})
	defer close(done)
	p := NewPipe()
	mux := http.NewServeMux()

	msgStream := Source(1000*time.Millisecond, 64, done)
	pipeline := p.DataDispatcher(2000*time.Millisecond, done, p.DataIntegrity(500*time.Millisecond, done, p.FileReceiver(1000*time.Millisecond, done, msgStream)))
	pipeline = p.FileSender(3000*time.Millisecond, done, p.DataArchiver(2000*time.Millisecond, done, p.FileMaker(2000*time.Millisecond, done, pipeline)))
	p.Sink(2000*time.Millisecond, done, pipeline)

	var port int
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.Parse()

	stripped, err := fs.Sub(frontend, "web/public")
	if err != nil {
		log.Fatalln(err)
	}

	frontendFS := http.FileServer(http.FS(stripped))
	mux.Handle("/", frontendFS)
	mux.HandleFunc("/api/v1/law", getRandomLaw)

	//http.HandleFunc("/api/v1/states/current/{id}", p.Update)
	mux.HandleFunc("/api/v1/states/current", p.UpdateState)

	mux.HandleFunc("/api/v1/ws", p.Notify)
	mux.HandleFunc("/api/v1/mon", p.Monitor)

	log.Printf("Server is listenig on localhost:%d\n", port)
	log.Fatalln(
		http.ListenAndServe(fmt.Sprintf("localhost:%d", port),
			handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
				handlers.AllowedOrigins([]string{"*"}),
			)(mux)))
	log.Println("<= main")
}

// Subscribe Create a regular, two-way channel but Returning it,
// implicitly converts it to a receive-only, as per the function return type.
// Subscribe adds a receive-only channel into a map entry, whose value is an array of receive-only channels.
// Subscribe return a pointer to this newly created (receive-only) channel.
// Notice by default a channel can both send and receive data, like func f(c chan string).
// But we can define a channel as func f(c <- chan string) to have a receive-only chan.
// Anyway if we want a send only chan, we create it as func f(c chan <- string).
func (p *Pipe) Subscribe(topic string) <-chan Message {
	log.Println("=> Subscribe")
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan Message, 1)
	p.subs[topic] = append(p.subs[topic], ch)
	log.Println("<= Subscribe")
	return ch
}

// Publish send a message to all of the channels registered at a given map entry.
func (p *Pipe) Publish(topic string, msg Message) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return
	}

	for _, ch := range p.subs[topic] {
		go func(ch chan Message) {
			log.Printf("Publish[%s]: %v\n", topic, msg)
			ch <- msg
		}(ch)
	}
}

func (p *Pipe) UpdateState(w http.ResponseWriter, r *http.Request) {
	log.Println("=> UpdateState")
	if r.Method != "PUT" {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var job Job
	json.NewDecoder(r.Body).Decode(&job)
	log.Printf("job: %v", job)
	json.NewEncoder(w).Encode(job)
	log.Println("<= UpdateState")
}

func StateMonitor() Message {
	log.Println("=> StateMonitor")
	rand.Seed(time.Now().UnixNano())
	var msg Message
	msgs := []Message{
		{
			Id:          1,
			State:       0,
			Description: "File receiver",
		},
		{
			Id:          2,
			State:       0,
			Description: "Data integrity",
		},
		{
			Id:          3,
			State:       0,
			Description: "Data dispatcher",
		},
		{
			Id:          4,
			State:       0,
			Description: "File maker",
		},
		{
			Id:          5,
			State:       0,
			Description: "Data archiving",
		},
		{
			Id:          6,
			State:       0,
			Description: "File sender",
		},
	}
	msg = msgs[rand.Intn(5)]
	time.Sleep(8 * time.Second)
	log.Printf("msg = %v", msg)
	log.Println("<= StateMonitor")
	return msg
}

func (p *Pipe) Monitor(w http.ResponseWriter, r *http.Request) {
	log.Println("=> Monitor")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		msg := StateMonitor()
		if err = c.WriteJSON(&msg); err != nil {
			log.Println("write:", err)
		}
		log.Println(msg)
	}
	//log.Println("<= Monitor")
}

// Notify extend the classical web server by sending a channel to the handler.
// There are two ways of doing this other than exporting the channels ...
// The first one is to use a function to return another handler function.
// When the function is returned, it will create a closure around the channel.
// The second is to use a struct which holds the channel as a member and use pointer receiver methods to handle the request.
func (p *Pipe) Notify(w http.ResponseWriter, r *http.Request) {
	log.Println("=> Notify")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	msgStream := p.Subscribe(RxEvent)
	// Waiting for messages coming for any publishers on that registered event key.
	for msg := range msgStream {
		if err = c.WriteJSON(&msg); err != nil {
			log.Println("write:", err)
		}
		log.Println(msg)
	}
	/*
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	*/
	log.Println("<= Notify")
}

// Source constructs a buffered channel of message types, with the length equal to the size parameter.
// Source starts a goroutine and return the constructed channel.
// Then on the gourontine created, Source sends up to size new messages on the channel it created.
// So in a nutshell, the Source  converts a discrete set of values into a stream of data on a channel.
// Most of time at the beginning of the pipeline, there is some batch of data one need to convert to a channel.
func Source(interval time.Duration, siz int, done <-chan interface{}) <-chan Message {
	msgStream := make(chan Message, 8)
	go func() {
		defer close(msgStream)

		for i := 1; i < siz+1; i++ {
			msg := Message{
				Id:          i,
				State:       0,
				Description: fmt.Sprintf("file-%03d", i),
			}
			select {
			case <-done:
				return
			case msgStream <- func(m Message) Message { log.Printf("Generator: %v\n", m); return m }(msg):
			}
			time.Sleep(interval)
		}
	}()
	return msgStream
}

// FileReceiver pipeline stage
func (p *Pipe) FileReceiver(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) <-chan Message {
	consumerStream := make(chan Message)
	go func() {
		defer close(consumerStream)
		for msg := range msgStream {
			select {
			case <-done:
				return
			case consumerStream <- func(m Message) Message {
				time.Sleep(interval)
				m.State = 1
				p.Publish(RxEvent, m)
				log.Printf("FileReceiver: %v\n", m)
				return m
			}(msg):
			}
		}
	}()
	return consumerStream
}

// DataIntegrity pipeline stage
func (p *Pipe) DataIntegrity(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) <-chan Message {
	workerStream := make(chan Message)
	go func() {
		defer close(workerStream)
		for msg := range msgStream {
			select {
			case <-done:
				return
			case workerStream <- func(m Message) Message {
				time.Sleep(interval)
				m.State = 2
				p.Publish(RxEvent, m)
				log.Printf("DataIntegrity: %v\n", m)
				return m
			}(msg):
			}
		}
	}()
	return workerStream
}

// DataIntegrity pipeline stage
func (p *Pipe) DataDispatcher(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) <-chan Message {
	workerStream := make(chan Message)
	go func() {
		defer close(workerStream)
		for msg := range msgStream {
			select {
			case <-done:
				return
			case workerStream <- func(m Message) Message {
				time.Sleep(interval)
				m.State = 3
				p.Publish(RxEvent, m)
				log.Printf("DataDispatcher: %v\n", m)
				return m
			}(msg):
			}
		}
	}()
	return workerStream
}

func (p *Pipe) FileMaker(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) <-chan Message {
	workerStream := make(chan Message)
	go func() {
		defer close(workerStream)
		for msg := range msgStream {
			select {
			case <-done:
				return
			case workerStream <- func(m Message) Message {
				time.Sleep(interval)
				m.State = 4
				p.Publish(RxEvent, m)
				log.Printf("FileMaker: %v\n", m)
				return m
			}(msg):
			}
		}
	}()
	return workerStream
}

func (p *Pipe) DataArchiver(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) <-chan Message {
	workerStream := make(chan Message)
	go func() {
		defer close(workerStream)
		for msg := range msgStream {
			select {
			case <-done:
				return
			case workerStream <- func(m Message) Message {
				time.Sleep(interval)
				m.State = 5
				p.Publish(RxEvent, m)
				log.Printf("DataArchiver: %v\n", m)
				return m
			}(msg):
			}
		}
	}()
	return workerStream
}

// FileSender pipeline stage
func (p *Pipe) FileSender(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) <-chan Message {
	producerStream := make(chan Message)
	go func() {
		defer close(producerStream)
		for msg := range msgStream {
			select {
			case <-done:
				return
			case producerStream <- func(m Message) Message {
				time.Sleep(interval)
				m.State = 6
				p.Publish(RxEvent, m)
				log.Printf("FileSender: %v\n", m)
				return m
			}(msg):
			}
		}
	}()
	return producerStream
}

// Sink is the final type of a pipeline function.
// It consumes input from prior stages but does not send output to subsequent pipeline stages.
func (p *Pipe) Sink(interval time.Duration, done <-chan interface{}, msgStream <-chan Message) {
	go func() {
		msgStream := p.Subscribe(RxEvent)
		for msg := range msgStream {
			log.Printf("Sink.Subscribe: %v\n", msg)
		}
	}()
	go func() {
		for msg := range msgStream {
			time.Sleep(interval)
			msg.State = 0
			p.Publish(RxEvent, msg)
			log.Printf("Sink.Pipe: %v\n", msg)
		}
	}()
}

type Law struct {
	Name       string `json:"name,omitempty"`
	Definition string `json:"definition,omitempty"`
}

var HackerLaws = []Law{
	{
		Name:       "Amdahl's Law",
		Definition: "Amdahl's Law is a formula which shows the potential speedup of a computational task which can be achieved by increasing the resources of a system.",
	},
	{
		Name:       "Conway's Law",
		Definition: "This law suggests that the technical boundaries of a system will reflect the structure of the organisation.",
	},
	{
		Name:       "Gall's Law",
		Definition: "A complex system that works is invariably found to have evolved from a simple system that worked.",
	},
}

func getRandomLaw(w http.ResponseWriter, r *http.Request) {
	log.Println("=> getRandomLaw")
	randomLaw := HackerLaws[rand.Intn(len(HackerLaws))]
	j, err := json.Marshal(randomLaw)
	if err != nil {
		http.Error(w, "couldn't retrieve random hacker law", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, bytes.NewReader(j))
	log.Println("<= getRandomLaw")
}
