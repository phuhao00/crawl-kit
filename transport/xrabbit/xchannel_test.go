package xrabbit

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func NewChannel(exchange, kind string) (*amqp.Channel, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", "guest", "guest", "127.0.0.1", 5672))
	if err != nil {
		return nil, nil
	}
	exchangeName := fmt.Sprintf("[PS]%s", exchange)
	ch, err := conn.Channel()
	err = ch.ExchangeDeclare(
		exchangeName, // name
		kind,         // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	return ch, err
}

func TestInActionP(t *testing.T) {

	channel,err:=NewChannel("exchange_test","fanout")
	if err != nil {
		fmt.Println(err)
	}
	ch:= NewXChannel(channel,"abc-queue")
	q := &amqp.Queue{Name: "xchannel-queue"}
	pCtx:=context.Background()
	ctx:=context.WithValue(pCtx,ContextKeyExchange,"exchange_test")
	ctx=context.WithValue(pCtx,ContextKeyPublishKey,"abc-key")
	pub := NewPublisher(
		ch,
		q,
		func(context.Context, *amqp.Publishing, interface{}) error { return nil },
		func(context.Context, *amqp.Delivery) (response interface{}, err error) {
			return struct{}{}, nil
		},
		PublisherTimeout(50*time.Second),
		PublisherDeliverer(ch.XChannelPubDeliver),
	)
	errChan := make(chan error, 1)
	go func() {
		for {
			select {
				case <-time.After(5 * time.Second):
					_, err := pub.Endpoint()(ctx, struct{}{})
					errChan <- err
			}
		}


	}()
	go func() {
		select {
		case err = <-errChan:
			fmt.Println(err)
		case <-time.After(100 * time.Second):
			t.Fatal("timed out waiting for result")
		}
		if err == nil {
			t.Error("expected error")
		}
	}()
	//if want, have := context.DeadlineExceeded.Error(), err.Error(); want != have {
	//	t.Errorf("want %s, have %s", want, have)
	//}
}
func TestInActionC(t *testing.T) {
	channel,err:=NewChannel("exchange_test","fanout")
	if err != nil {
		fmt.Println(err)
	}
	ch:= NewXChannel(channel,"abc-queue")
	q := &amqp.Queue{Name: "xchannel-queue"}
	pCtx:=context.Background()
	ctx:=context.WithValue(pCtx,ContextKeyExchange,"exchange_test")
	ctx=context.WithValue(pCtx,ContextKeyPublishKey,"abc-key")
	pub := NewPublisher(
		ch,
		q,
		func(context.Context, *amqp.Publishing, interface{}) error { return nil },
		func(context.Context, *amqp.Delivery) (response interface{}, err error) {
			return struct{}{}, nil
		},
		PublisherTimeout(50*time.Second),
		PublisherDeliverer(ch.XChannelConDeliver),
	)
	resChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-time.After(5 * time.Second):
				res, err := pub.Endpoint()(ctx, struct{}{})
				if err != nil {
					errChan <- err
				} else {
					resChan <- res
				}
			}
		}
	}()

	go func() {
		for  {
			select {
			case response := <-resChan:
				fmt.Println(response)
			case err = <-errChan:
				fmt.Println(err)
			case <-time.After(100 * time.Second):
				t.Fatal("timed out waiting for result")
			}
			if err == nil {
				t.Error("expected error")
			}
		}
	}()
	//if want, have := context.DeadlineExceeded.Error(), err.Error(); want != have {
	//	t.Errorf("want %s, have %s", want, have)
	//}
}
