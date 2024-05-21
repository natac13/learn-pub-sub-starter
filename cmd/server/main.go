package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connString := "amqp://guest:guest@localhost:5672/"

	conn, err := amqp.Dial(connString)
	if err != nil {
		slog.Error("could not connect to RabbitMQ", "error", err)
		os.Exit((1))
	}
	defer func() {
		slog.Info("closing connection to RabbitMQ")
		conn.Close()
	}()

	fmt.Println("\n nowconnected to RabbitMQ")

	// create a channel
	ch, err := conn.Channel()
	if err != nil {
		slog.Error("could not create channel", "error", err)
		os.Exit(1)
	}

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerLogs(),
	)

	if err != nil {
		slog.Error("could not declare and bind queue", "error", err)
		os.Exit(1)
	}

	// fmt.Println("queue declared and bound", q.Name)

	// wait for ctrl+c
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt)

	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		if strings.HasPrefix(words[0], "pause") {
			fmt.Println("pausing...")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			continue
		}

		if strings.HasPrefix(words[0], "resume") {
			fmt.Println("resuming...")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			continue
		}

		if strings.HasPrefix(words[0], "quit") {
			fmt.Println("shutting down...")
			return
		}

		fmt.Println("unknown command")

	}

}
func handlerLogs() func(gamelog routing.GameLog) pubsub.Acktype {
	return func(gamelog routing.GameLog) pubsub.Acktype {
		defer fmt.Print("> ")

		err := gamelogic.WriteLog(gamelog)
		if err != nil {
			fmt.Printf("error writing log: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
