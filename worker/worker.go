package worker

import (
	"encoding/json"
	"strconv"

	"github.com/assembla/cony"
	"github.com/mono83/slf/wd"
	"github.com/streadway/amqp"

	"elyby/minecraft-skinsystem/db"
	"elyby/minecraft-skinsystem/interfaces"
	"elyby/minecraft-skinsystem/model"
)

type Services struct {
	AmqpClient  *cony.Client
	SkinsRepo   interfaces.SkinsRepository
	AccountsAPI interfaces.AccountsAPI
	Logger      wd.Watchdog
}

type UsernameChanged struct {
	AccountId   int    `json:"accountId"`
	OldUsername string `json:"oldUsername"`
	NewUsername string `json:"newUsername"`
}

type SkinChanged struct {
	AccountId       int    `json:"userId"`
	Uuid            string `json:"uuid"`
	SkinId          int    `json:"skinId"`
	Hash            string `json:"hash"`
	Is1_8           bool   `json:"is1_8"`
	IsSlim          bool   `json:"isSlim"`
	Url             string `json:"url"`
	MojangTextures  string `json:"mojangTextures"`
	MojangSignature string `json:"mojangSignature"`
}

const exchangeName string = "events"
const queueName string = "skinsystem-accounts-events"

func (service *Services) Run() error {
	clientErrs, consumerErrs, deliveryChannel := setupClient(service.AmqpClient)
	shouldReturnError := true

	for service.AmqpClient.Loop() {
		select {
		case msg := <-deliveryChannel:
			shouldReturnError = false
			service.HandleDelivery(&msg)
		case err := <-consumerErrs:
			if shouldReturnError {
				return err
			}

			service.Logger.Error("Consume error: :err", wd.ErrParam(err))
		case err := <-clientErrs:
			if shouldReturnError {
				return err
			}

			service.Logger.Error("Client error: :err", wd.ErrParam(err))
		}
	}

	return nil
}

func (service *Services) HandleDelivery(delivery *amqp.Delivery) {
	service.Logger.Debug("Incoming message with routing key " + delivery.RoutingKey)
	var result bool = true
	switch delivery.RoutingKey {
	case "accounts.username-changed":
		var event *UsernameChanged
		json.Unmarshal(delivery.Body, &event)
		result = service.HandleChangeUsername(event)
	case "accounts.skin-changed":
		var event *SkinChanged
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

func (service *Services) HandleChangeUsername(event *UsernameChanged) bool {
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
func (service *Services) HandleSkinChanged(event *SkinChanged) bool {
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

func setupClient(client *cony.Client) (<-chan error, <-chan error, <-chan amqp.Delivery ) {
	exchange := cony.Exchange{
		Name:       exchangeName,
		Kind:       "topic",
		Durable:    true,
		AutoDelete: false,
	}

	queue := &cony.Queue{
		Name:       queueName,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
	}

	usernameEventBinding := cony.Binding{
		Exchange: exchange,
		Queue:    queue,
		Key:      "accounts.username-changed",
	}

	skinEventBinding := cony.Binding{
		Exchange: exchange,
		Queue:    queue,
		Key:      "accounts.skin-changed",
	}

	declarations := []cony.Declaration{
		cony.DeclareExchange(exchange),
		cony.DeclareQueue(queue),
		cony.DeclareBinding(usernameEventBinding),
		cony.DeclareBinding(skinEventBinding),
	}

	client.Declare(declarations)

	consumer := cony.NewConsumer(queue,
		cony.Qos(10),
		cony.AutoTag(),
	)
	client.Consume(consumer)

	return client.Errors(), consumer.Errors(), consumer.Deliveries()
}
