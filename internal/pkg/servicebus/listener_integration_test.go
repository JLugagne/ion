package servicebus

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"pack.ag/amqp"

	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"
)

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

var config = &types.Configuration{
	ClientID:            os.Getenv("AZURE_CLIENT_ID"),
	ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
	ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
	SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
	TenantID:            os.Getenv("AZURE_TENANT_ID"),
	ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
	Hostname:            "Test",
	ModuleName:          helpers.RandomName(8),
	SubscribesToEvent:   "ExampleEvent2",
	EventsPublished:     "ExamplePublishtopic",
	LogLevel:            "Debug",
	Job: &types.JobConfig{
		RetryCount: 5,
	},
}

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationNewListener(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	listener := NewListener(ctx, config)
	// Remove topic to ensure each test has a clean topic to work with
	defer deleteSubscription(listener, config)

	nonce := time.Now().String()
	sender := createAmqpSender(listener)
	err := sender.Send(ctx, &amqp.Message{
		Value: nonce,
	})
	if err != nil {
		t.Error(err)
	}

	depth, err := listener.GetQueueDepth()
	if err != nil || depth == nil {
		t.Error("Failed to get queue depth")
		t.Error(err)
	}

	derefDepth := *depth

	if derefDepth != 1 {
		t.Errorf("Expected queue depth of 1 Got:%v", derefDepth)
		t.Fail()
	}

	amqpMessage, err := listener.AmqpReceiver.Receive(ctx)
	if err != nil {
		t.Error(err)
	}

	message := messaging.NewAmqpMessageWrapper(amqpMessage)

	message.Accept()
	if message.Body().(string) != nonce {
		t.Errorf("value not as expected in message Expected: %s Got: %s", nonce, message.Body())
	}

	depth, err = listener.GetQueueDepth()
	if err != nil || depth == nil {
		t.Error("Failed to get queue depth")
		t.Error(err)
	}

	derefDepth = *depth

	if derefDepth != 0 {
		t.Errorf("Expected queue depth of 0 Got:%v", derefDepth)
		t.Fail()
	}
}

func TestIntegrationRequeueReleasedMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	listener := NewListener(ctx, config)
	// Remove topic to ensure each test has a clean topic to work with
	defer deleteSubscription(listener, config)

	nonce := time.Now().String()
	sender := createAmqpSender(listener)
	err := sender.Send(ctx, &amqp.Message{
		Value: nonce,
	})
	if err != nil {
		t.Error(err)
	}

	for index := 0; index < 6; index++ {
		amqpMessage, err := listener.AmqpReceiver.Receive(ctx)
		message := messaging.NewAmqpMessageWrapper(amqpMessage)
		if err != nil {
			t.Error(err)
		}

		if message.DeliveryCount() != index {
			t.Logf("Delivery count: Got %v Expected %v", message.DeliveryCount(), index)
		}

		err = message.Reject()
		if err != nil {
			t.Error(err)
		}
	}

	checkUntil := time.Now().Add(time.Second * 3)
	checkCtx, cancel := context.WithDeadline(context.Background(), checkUntil)
	defer cancel()

	_, err = listener.AmqpReceiver.Receive(checkCtx)
	if err != nil {
		t.Log(err)
	} else {
		t.Error("message delivered a 6th time - after 5 should be deadlettered")
	}
}

// createAmqpSender exists for e2e testing.
func createAmqpSender(listener *Listener) *amqp.Sender {
	if listener.AmqpSession == nil {
		log.WithField("currentListener", listener).Panic("Cannot create amqp listener without a session already configured")
	}

	sender, err := listener.AmqpSession.NewSender(
		amqp.LinkTargetAddress("/" + listener.TopicName),
	)
	if err != nil {
		log.Fatal("Creating receiver:", err)
	}

	return sender
}

func deleteSubscription(listener *Listener, config *types.Configuration) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*45)
	defer cancel()
	_, err := listener.subsClient.Delete(ctx, config.ResourceGroup, config.ServiceBusNamespace, listener.TopicName, listener.SubscriptionName)
	if err != nil {
		panic(err)
	}
}