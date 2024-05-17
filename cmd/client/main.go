package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

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

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		slog.Error("could not get username", "error", err)
		os.Exit(1)
	}

	_, q, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, "pause."+username, routing.PauseKey, pubsub.SimpleQueueTypeTransient)

	if err != nil {
		slog.Error("could not declare and bind queue", "error", err)
		os.Exit(1)
	}

	fmt.Println("queue declared and bound", q.Name)

	// wait for ctrl+c
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt)

	for {
		select {
		case <-shutdownChan:
			fmt.Println("shutting down...")
			return
		}
	}

}
