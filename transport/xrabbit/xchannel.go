package xrabbit

import (
	"context"
	"errors"
	"github.com/streadway/amqp"
)

type XChannel struct {
	channel       *amqp.Channel
	IsConsumeInit bool
	queueName     string
}

func NewXChannel(ch*amqp.Channel,queueName string) *XChannel {
	return &XChannel{
		channel:       ch,
		IsConsumeInit: false,
		queueName:     queueName,
	}
}

func (c XChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return c.channel.Publish(exchange,key,mandatory,immediate,msg)
}
func (c XChannel)Consume(queue, consumer string, autoAck, exclusive, noLocal, noWail bool, args amqp.Table) (<-chan amqp.Delivery, error){
	return c.channel.Consume(queue,consumer,autoAck,exclusive,noLocal,noWail,args)
}

func (c *XChannel) PublishEntrance(ctx context.Context, Publishing amqp.Publishing) error {
	if err := c.Publish(
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

func (c *XChannel) ConsumeEntrance(ctx context.Context) (*amqp.Delivery, error) {
	if !c.IsConsumeInit {
		return nil, errors.New("channel for Consume  not init ")
	}
	autoAck := getConsumeAutoAck(ctx)

	msg, err := c.Consume(
		c.queueName,
		getConsumerName(ctx), //consumer
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

func (c *XChannel)XChannelPubDeliver(ctx context.Context,
	p Publisher,
	pub *amqp.Publishing) (*amqp.Delivery, error)  {
	err := c.PublishEntrance(
		ctx,
		*pub,
	)
	return nil, err
}

func (c *XChannel)XChannelConDeliver(ctx context.Context,
	p Publisher,
	pub *amqp.Publishing) (*amqp.Delivery, error)  {
	return  c.ConsumeEntrance(ctx)
}
