// The buffered write is much faster than the unbeffered.
// This is because in bufio.writer, the writes are queued internally into a buffer and then written out as a single chunk of data.
package buf_test

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

// GOMAXPROCS  controls the number of operating system threads allocated to goroutines.
func BenchmarkUnbufferedWrite(b *testing.B) {
	performWrite(b, tmpFileOrFatal())
}

func BenchmarkBufferedWrite(b *testing.B) {
	bufferredFile := bufio.NewWriter(tmpFileOrFatal())
	performWrite(b, bufio.NewWriter(bufferredFile))
}

func tmpFileOrFatal() *os.File {
	file, err := os.CreateTemp("", "tmp")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return file
}

// Below are two handy generators.
// - repeat is a generator that will repeat the values we pass to it infinitely until we tell it to stop.
// - take will only take the first num items of its incoming valueStream and the exit.
func performWrite(b *testing.B, writer io.Writer) {
	repeat := func(
		done <-chan interface{},
		values ...interface{},
	) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}
	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()
		return takeStream
	}

	done := make(chan interface{})
	defer close(done)

	b.ResetTimer()
	// Notice this pipeline stage will only take the first b.N bytes off its incoming valueStream and the exit.
	fmt.Printf("Number of times we perform the op being mesured: [%d]\n", b.N)
	for bt := range take(done, repeat(done, byte(0)), b.N) {
		writer.Write([]byte{bt.(byte)})
	}
}
