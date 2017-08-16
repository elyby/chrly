package worker

import (
	"encoding/json"

	"github.com/mono83/slf/wd"
	"github.com/streadway/amqp"

	"elyby/minecraft-skinsystem/model"
	"elyby/minecraft-skinsystem/repositories"
)

type Services struct {
	Channel   *amqp.Channel
	SkinsRepo repositories.SkinsRepository
	Logger    wd.Watchdog
}

const exchangeName string = "events"
const queueName string = "skinsystem-accounts-events"

func (service *Services) Run() error {
	deliveryChannel, err := setupConsume(service.Channel)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range deliveryChannel {
			service.Logger.Debug("Incoming message with routing key " + d.RoutingKey)
			var result bool = true
			switch d.RoutingKey {
			case "accounts.username-changed":
				var event *model.UsernameChanged
				json.Unmarshal(d.Body, &event)
				result = service.HandleChangeUsername(event)
			case "accounts.skin-changed":
				var event *model.SkinChanged
				json.Unmarshal(d.Body, &event)
				result = service.HandleSkinChanged(event)
			}

			if result {
				d.Ack(false)
			} else {
				d.Reject(true)
			}
		}
	}()
	<-forever

	return nil
}

func (service *Services) HandleChangeUsername(event *model.UsernameChanged) bool {
	if event.OldUsername == "" {
		service.Logger.IncCounter("worker.change_username.empty_old_username", 1)
		record := &model.Skin{
			UserId:   event.AccountId,
			Username: event.NewUsername,
		}

		service.SkinsRepo.Save(record)

		return true
	}

	record, err := service.SkinsRepo.FindByUserId(event.AccountId)
	if err != nil {
		/*
		// TODO: вернуть логику восстановления информации об аккаунте
		service.Logger.IncCounter("worker.change_username.id_not_found", 1)
		service.Logger.Warning("Cannot find user id. Trying to search.")
		response, err := getById(event.AccountId)
		if err != nil {
			service.Logger.IncCounter("worker.change_username.id_not_restored", 1)
			service.Logger.Error("Cannot restore user info. %T\n", err)
			// TODO: логгировать в какой-нибудь Sentry, если там не 404
			return true
		}

		service.Logger.IncCounter("worker.change_username.id_restored", 1)
		fmt.Println("User info successfully restored.")
		record = &event.Skin{
			UserId: response.Id,
		}
		*/
	}

	record.Username = event.NewUsername
	service.SkinsRepo.Save(record)

	service.Logger.IncCounter("worker.change_username.processed", 1)

	return true
}

func (service *Services) HandleSkinChanged(event *model.SkinChanged) bool {
	record, err := service.SkinsRepo.FindByUserId(event.AccountId)
	if err != nil {
		service.Logger.IncCounter("worker.skin_changed.id_not_found", 1)
		service.Logger.Warning("Cannot find user id. Trying to search.")
		/*
		// TODO: вернуть логику восстановления информации об аккаунте
		response, err := getById(event.AccountId)
		if err != nil {
			services.Logger.IncCounter("worker.skin_changed.id_not_restored", 1)
			fmt.Printf("Cannot restore user info. %T\n", err)
			// TODO: логгировать в какой-нибудь Sentry, если там не 404
			return true
		}

		services.Logger.IncCounter("worker.skin_changed.id_restored", 1)
		fmt.Println("User info successfully restored.")
		record.UserId = response.Id
		record.Username = response.Username
		*/
	}

	record.Uuid = event.Uuid
	record.SkinId = event.SkinId
	record.Hash = event.Hash
	record.Is1_8 = event.Is1_8
	record.IsSlim = event.IsSlim
	record.Url = event.Url
	record.MojangTextures = event.MojangTextures
	record.MojangSignature = event.MojangSignature

	service.SkinsRepo.Save(record)

	service.Logger.IncCounter("worker.skin_changed.processed", 1)

	return true
}

func setupConsume(channel *amqp.Channel) (<-chan amqp.Delivery, error) {
	var err error
	err = channel.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, err
	}

	_, err = channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(queueName, "accounts.username-changed", exchangeName, false, nil)
	if err != nil {
		return nil, err
	}

	err = channel.QueueBind(queueName, "accounts.skin-changed", exchangeName, false, nil)
	if err != nil {
		return nil, err
	}

	deliveryChannel, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return nil, err
	}

	return deliveryChannel, nil
}
