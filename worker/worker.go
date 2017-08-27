package worker

import (
	"encoding/json"
	"strconv"

	"github.com/mono83/slf/wd"
	"github.com/streadway/amqp"

	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/interfaces"
	"elyby/minecraft-skinsystem/model"
)

type Services struct {
	Channel     *amqp.Channel
	SkinsRepo   interfaces.SkinsRepository
	AccountsAPI interfaces.AccountsAPI
	Logger      wd.Watchdog
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
			service.HandleDelivery(&d)
		}
	}()
	<-forever

	return nil
}

func (service *Services) HandleDelivery(delivery *amqp.Delivery) {
	service.Logger.Debug("Incoming message with routing key " + delivery.RoutingKey)
	var result bool = true
	switch delivery.RoutingKey {
	case "accounts.username-changed":
		var event *model.UsernameChanged
		json.Unmarshal(delivery.Body, &event)
		result = service.HandleChangeUsername(event)
	case "accounts.skin-changed":
		var event *model.SkinChanged
		json.Unmarshal(delivery.Body, &event)
		result = service.HandleSkinChanged(event)
	default:
		service.Logger.Info("Unknown delivery with routing key " + delivery.RoutingKey)
		delivery.Ack(false)
		return
	}

	if result {
		delivery.Ack(false)
	} else {
		delivery.Reject(true)
	}
}

func (service *Services) HandleChangeUsername(event *model.UsernameChanged) bool {
	service.Logger.IncCounter("worker.change_username", 1)
	if event.OldUsername == "" {
		service.Logger.IncCounter("worker.change_username_empty_old_username", 1)
		record := &model.Skin{
			UserId:   event.AccountId,
			Username: event.NewUsername,
		}

		service.SkinsRepo.Save(record)

		return true
	}

	record, err := service.SkinsRepo.FindByUserId(event.AccountId)
	if err != nil {
		service.Logger.Info("Cannot find user id :accountId. Trying to search.", wd.IntParam("accountId", event.AccountId))
		if _, isSkinNotFound := err.(*db.SkinNotFoundError); !isSkinNotFound {
			service.Logger.Error("Unknown error when requesting a skin from the repository: :err", wd.ErrParam(err))
		}

		service.Logger.IncCounter("worker.change_username_id_not_found", 1)
		record = &model.Skin{
			UserId: event.AccountId,
		}
	}

	record.Username = event.NewUsername
	service.SkinsRepo.Save(record)

	return true
}

// TODO: возможно стоит добавить проверку на совпадение id аккаунтов
func (service *Services) HandleSkinChanged(event *model.SkinChanged) bool {
	service.Logger.IncCounter("worker.skin_changed", 1)
	var record *model.Skin
	record, err := service.SkinsRepo.FindByUserId(event.AccountId)
	if err != nil {
		if _, isSkinNotFound := err.(*db.SkinNotFoundError); !isSkinNotFound {
			service.Logger.Error("Unknown error when requesting a skin from the repository: :err", wd.ErrParam(err))
		}

		service.Logger.IncCounter("worker.skin_changed_id_not_found", 1)
		service.Logger.Info("Cannot find user id :accountId. Trying to search.", wd.IntParam("accountId", event.AccountId))
		response, err := service.AccountsAPI.AccountInfo("id", strconv.Itoa(event.AccountId))
		if err != nil {
			service.Logger.IncCounter("worker.skin_changed_id_not_restored", 1)
			service.Logger.Error(
				"Cannot restore user info for :accountId: :err",
				wd.IntParam("accountId", event.AccountId),
				wd.ErrParam(err),
			)

			return true
		}

		service.Logger.IncCounter("worker.skin_changed_id_restored", 1)
		service.Logger.Info("User info successfully restored.")

		record = &model.Skin{
			UserId: response.Id,
			Username: response.Username,
		}
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
