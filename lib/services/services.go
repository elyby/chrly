package services

import (
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/streadway/amqp"
)

var RedisPool *pool.Pool

var RabbitMQChannel *amqp.Channel
