package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/iamalsaher/interactor/pkg/process"
)

var wg sync.WaitGroup

func interactor(input io.Writer, output io.Reader) {
	var b bytes.Buffer

	b.Grow(8192)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				if b.Len() > 0 {
					fmt.Print(string(b.Next(b.Len())))
				}
			}
		}
	}()

	io.Copy(&b, output)
	wg.Done()
	<-done

}

func main() {
	proc := process.NewProcess("./binary", "--sleep", "5", "-env")
	proc.SetEnviron([]string{"SEXYENV=LOL"}, true)
	// proc.SetDirectory("/tmp")
	// proc.SetTimeout(1000)
	if e := proc.ConnectIO(false, false); e != nil {
		panic(e)
	}

	wg.Add(1)
	if e := proc.Start(&process.Interactor{Function: interactor, Input: proc.Pipe.StdinW, Output: proc.Pipe.StdoutR}); e != nil {
		panic(e)
	}

	fmt.Printf("Started Process with PID %v\n", proc.PID)

	ps, err := proc.Wait()
	if err != nil {
		log.Fatalf("Error waiting for process: %v", err)
	}
	wg.Wait()
	fmt.Printf("exit code was: %d\n", ps.ExitCode())
}
