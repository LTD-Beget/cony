package cony_test

import (
	"log"
	"os"
	"time"

	"github.com/LTD-Beget/cony"
	"github.com/streadway/amqp"
)

func Example() {
	client := cony.NewClient(cony.URL(os.Getenv("AMQP_URL")), cony.Backoff(cony.DefaultBackoff))

	q := &cony.Queue{
		Name:       "", // autogenerated queue name
		AutoDelete: true,
	}

	exchange := cony.Exchange{
		Name:    "amq.topic",
		Durable: true,
	}

	b := cony.Binding{
		Queue:    q,
		Exchange: exchange,
		Key:      "something.#",
	}

	// wrap all declarations and save into slice
	declarations := []cony.Declaration{
		cony.DeclareQueue(q),
		cony.DeclareExchange(exchange),
		cony.DeclareBinding(b),
	}

	// declare consumer
	consumer := cony.NewConsumer(q,
		cony.Qos(10),
		cony.AutoTag(),
		cony.AutoAck(),
	)

	// declare publisher
	publisher := cony.NewPublisher(exchange.Name,
		"ololo.key",
		cony.PublishingTemplate(amqp.Publishing{
			ContentType: "application/json",
			AppId:       "app1",
		}), // template amqp.Publising
	)

	// let client know about declarations
	client.Declare(declarations)

	// let client know about consumers/publishers
	client.Consume(consumer)
	client.Publish(publisher)

	clientErrs := client.Errors()
	deliveries := consumer.Deliveries()
	consumerErrs := consumer.Errors()

	// connect, reconnect, or exit loop
	// run network operations such as:
	// queue, exchange, bidning, consumers declarations
	for client.Loop() {
		select {
		case msg := <-deliveries:
			log.Println(msg)
			msg.Ack(false)
			publisher.Write([]byte("ololo reply"))
		case err := <-consumerErrs:
			log.Println("CONSUMER ERROR: ", err)
		case err := <-clientErrs:
			log.Println("CLIENT ERROR: ", err)
			client.Close()
		}
	}

}

func ExampleURL() {
	cony.NewClient(cony.URL("amqp://guest:guest@localhost/"))
}

func ExampleErrorsChan() {
	errors := make(chan error, 100) // define custom buffer size
	cony.NewClient(cony.ErrorsChan(errors))
}

func ExampleBlockingChan() {
	blockings := make(chan amqp.Blocking, 100) // define custom buffer size
	cony.NewClient(cony.BlockingChan(blockings))
}

func ExampleClient_Loop() {
	client := cony.NewClient(cony.URL("amqp://guest:guest@localhost/"))

	for client.Loop() {
		select {
		case err := <-client.Errors():
			log.Println("CLIENT ERROR: ", err)
			client.Close()
		}
		time.Sleep(1 * time.Second) // naive backoff
	}
}
