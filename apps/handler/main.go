package main

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

func main() {
	consume()
}

func consume() {
	var rabbit_host = os.Getenv("RABBIT_HOST")
	var rabbit_port = os.Getenv("RABBIT_PORT")
	var rabbit_user = os.Getenv("RABBIT_USERNAME")
	var rabbitmqPasswdFilename = os.Getenv("RABBIT_PASSWD_FILENAME")
	rabbit_password, err := os.ReadFile(rabbitmqPasswdFilename)
	if err != nil {
		log.Fatalf("reading rabbitmq password from file %s: %s", rabbitmqPasswdFilename, err)
	}

	var address = "amqp://" + rabbit_user + ":" + string(rabbit_password) + "@" + rabbit_host + ":" + rabbit_port + "/"
	conn, err := amqp.Dial(address)
	if err != nil {
		log.Fatalf("%s: %s %s", "Failed to connect to RabbitMQ", address, err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		os.Getenv("REDEEMED_QUEUE_NAME"), // name
		true,                             // durable
		false,                            // delete when unused
		false,                            // exclusive
		false,                            // no-wait
		nil,                              // arguments
	)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("%s: %s", "Failed to register consumer", err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("The Handler Received a message: %s!!!", d.Body)
			d.Ack(false)
		}
	}()
	log.Printf("Running...")
	fmt.Println("Running...")
	<-forever
}
