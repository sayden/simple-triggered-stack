package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sayden/simple-triggered-stack"
	"github.com/timjchin/unpuzzled"
)

type config struct {
	port string
}

var conf config

func main() {
	app := unpuzzled.NewApp()

	app.Name = "Server"
	app.Description = "Server"
	app.Usage = "See --help for info"
	app.Authors = []unpuzzled.Author{{Name: "Mario Castro", Email: "mariocaster@gmail.com"}}

	app.Command = &unpuzzled.Command{
		Name:  "Main",
		Usage: "Launches the server",
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "port",
				Destination: &conf.port,
				Default:     "8080",
				Description: "Server port",
			},
		},
		Action: launch,
	}
	app.Run(os.Args)
}

type enqueuableString struct {
	data string
}

func (e *enqueuableString) Data() interface{} {
	return e.data
}

func launch() {
	msgN := 10

	//Channel that the stack will use to notify us that flushing has complete
	quit := make(chan struct{})

	stackConfig := simple_triggered_stack.Config{
		MaxStack:           3,
		MaxIngestionBuffer: 10,
	}

	stackCallback := func(stack []simple_triggered_stack.Stackable) {
		for _, v := range stack {
			fmt.Printf("%s ", v.Data())
		}

		fmt.Println()
	}

	q := simple_triggered_stack.NewStack(quit, stackConfig, stackCallback)

	for i := 0; i < msgN; i++ {
		time.Sleep(time.Millisecond * 100)
		q.IngestionCh() <- &enqueuableString{fmt.Sprintf("Hello %d", i)}
	}

	//We have finished pushing, close ingestion channel. Queue will notify of successful flushing by closing quit ch
	close(q.IngestionCh())

	<-quit
}
