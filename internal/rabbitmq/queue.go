package rabbitmq

import (
	"context"

	"github.com/streadway/amqp"
	"github.com/vkuksa/twatter/internal"
)

const (
	consumerName = "queue"
)

type Queue struct {
	connection *amqp.Connection
	chanel     *amqp.Channel
	name       string
}

// Creates an instance of a Queue, that connects to rabbitmq, and declares a Queue
// url param specifies url to dial to rabbitmq server and name specifies queue name
func NewQueue(url, n string) (*Queue, error) {
	cn, err := amqp.Dial(url)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "amqp.Dial")
	}

	ch, err := cn.Channel()
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "cn.Channel")
	}

	_, err = ch.QueueDeclare(n, false, false, false, false, nil)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "ch.QueueDeclare")
	}

	return &Queue{connection: cn, chanel: ch, name: n}, nil
}

func (q *Queue) Close() {
	q.connection.Close()
}

func (q *Queue) Enqueue(_ context.Context, msg string) error {
	err := q.chanel.Publish(
		"",     // exchange
		q.name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "ch.Publish")
	}

	return nil
}

// func (q *Queue) Consume(ctx context.Context) (chan string, error) {
// msgs, err := q.chanel.Consume(
// 	q.name,       // queue
// 	consumerName, // consumer
// 	false,        // auto-ack
// 	false,        // exclusive
// 	false,        // no-local
// 	false,        // no-wait
// 	nil,          // args
// )
// if err != nil {
// 	return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "queue.Consume")
// }

// resChan := make(chan string)
// select {
// case <-ctx.Done():
// 	return nil, errors.New("Context done") //TODO: How to handle it?
// case msg := <-msgs:
// 	return string(msg.Body), nil
// }
// }
