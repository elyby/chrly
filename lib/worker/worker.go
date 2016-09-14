package worker

import (
	"log"

	"encoding/json"

	"elyby/minecraft-skinsystem/lib/services"
)

const exchangeName string = "events"
const queueName string = "skinsystem-accounts-events"

func Listen() {
	var err error
	ch := services.RabbitMQChannel

	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(queueName, "accounts.username-changed", exchangeName, false, nil)
	failOnError(err, "Failed to bind a queue")

	err = ch.QueueBind(queueName, "accounts.skin-changed", exchangeName, false, nil)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Println("Incoming message with routing key " + d.RoutingKey)
			var result bool = true;
			switch d.RoutingKey {
			case "accounts.username-changed":
				var model usernameChanged
				json.Unmarshal(d.Body, &model)
				result = handleChangeUsername(model)
			case "accounts.skin-changed":
				var model skinChanged
				json.Unmarshal(d.Body, &model)
				result = handleSkinChanged(model)
			}

			if (result) {
				d.Ack(false)
			} else {
				d.Reject(true)
			}
		}
	}()

	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
