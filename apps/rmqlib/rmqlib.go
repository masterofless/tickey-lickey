package rmqlib

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

func EnqueueMessage(message []byte) {
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
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		log.Fatalf("%s: %s", "Failed to publish a message", err)
	}
	log.Printf("publish message success %s!", message)
}
