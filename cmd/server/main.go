package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

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
