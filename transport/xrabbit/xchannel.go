package xrabbit

import (
	"context"
	"errors"
	"github.com/streadway/amqp"
)

type XChannel struct {
	Publisher 	*Publisher
	channel       *amqp.Channel
	IsConsumeInit bool
	queueName     string
}

func NewXChannel(
	ch Channel,
	q *amqp.Queue,
	enc EncodeRequestFunc,
	dec DecodeResponseFunc,
	options ...PublisherOption,
	) *XChannel {
	return &XChannel{
		Publisher: NewPublisher(
			ch,q,enc,dec,options...
			),
		channel:       nil,
		IsConsumeInit: false,
		queueName:     "",
	}
}

func (c *XChannel) Publish(ctx context.Context, Publishing amqp.Publishing) error {
	if err := c.channel.Publish(
		getPublishExchange(ctx), // exchange
		getPublishKey(ctx),      // routing key
		false,                   // mandatory
		false,                   // immediate
		Publishing,
	); err != nil {
		return err
	}
	return nil
}

func (c *XChannel) InitConsumeChannel(ctx context.Context) error {
	q, err := c.channel.QueueDeclare(
		getQueueName(ctx), // name
		false,             // durable
		false,             // delete when unused
		true,              // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		q.Name,                  // queue name
		getPublishKey(ctx),      // routing key
		getPublishExchange(ctx), // exchange
		false,
		getConsumeArgs(ctx))

	if err == nil {
		c.IsConsumeInit = true
	}
	return err
}

func (c *XChannel) Consume(ctx context.Context) (*amqp.Delivery, error) {
	if !c.IsConsumeInit {
		return nil, errors.New("channel for Consume  not init ")
	}
	autoAck := getConsumeAutoAck(ctx)

	msg, err := c.channel.Consume(
		c.Publisher.q.Name,
		"", //consumer
		autoAck,
		false, //exclusive
		false, //noLocal
		false, //noWait
		getConsumeArgs(ctx),
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case d := <-msg:
			if d.CorrelationId == "CorrelationId" {
				if !autoAck {
					d.Ack(false) //multiple
				}
				return &d, nil
			}

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
