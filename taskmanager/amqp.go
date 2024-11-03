package taskmanager

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/onedotnet/asynctasks/config"

	"github.com/streadway/amqp"
)

const (
	ExchangeDirect  = "direct"
	ExchangeFanout  = "fanout"
	ExchangeHeaders = "headers"
	ExchangeMatch   = "match"
	ExchangeTrace   = "rabbitmq.trace"
	ExchangeTopic   = "topic"
)

// QueueProvider 结构
type QueueProvider struct {
	conn          *amqp.Connection
	channel       *amqp.Channel
	connNotify    chan *amqp.Error
	channelNotify chan *amqp.Error
	quit          chan struct{}
	addr          string
	port          int
	username      string
	password      string
	vhost         string
	exchange      string
	exchangeType  string
	queue         string
	routingKey    string
	tag           string
	autoDelete    bool
	handler       func([]byte) error
	qos           int
	maxsize       int
	args          map[string]interface{}
	channelPool   chan *amqp.Channel
}

// NewQueueProvider 返回一个新的队列结构
func NewQueueProvider(exchange, exchangeKind, route, queue string, autoDelete bool, handler func([]byte) error) *QueueProvider {
	qp := &QueueProvider{
		username:     config.AppConfig.RMQUser,
		password:     config.AppConfig.RMQPass,
		vhost:        config.AppConfig.RMQVHost,
		addr:         config.AppConfig.RMQHost,
		port:         config.AppConfig.RMQPort,
		exchange:     exchange,
		exchangeType: exchangeKind,
		routingKey:   route,
		queue:        queue,
		tag:          "",
		autoDelete:   autoDelete,
		handler:      handler,
		quit:         make(chan struct{}),
		qos:          0,
		maxsize:      0,
		args:         nil,
	}
	channelPoolSize := 3
	if config.AppConfig.RMQChannelPoolSize > 0 {
		channelPoolSize = config.AppConfig.RMQChannelPoolSize
	}
	qp.channelPool = make(chan *amqp.Channel, channelPoolSize)
	return qp
}

// SetArgs 设置参数
func (q *QueueProvider) GetQueueRoute() string {
	return q.routingKey
}

// SetArgs 设置参数
func (q *QueueProvider) SetArgs(args map[string]interface{}) {
	q.args = args
}

// SetQOS 设置最大接收数量
func (q *QueueProvider) SetQOS(qos int) {
	q.qos = qos
}

// Start 启动一个队列
func (q *QueueProvider) Start() error {
	if err := q.Run(); err != nil {
		return err
	}
	go q.ReConnect()
	return nil
}

// Stop 停止一个队列
func (q *QueueProvider) Stop() {
	close(q.quit)

	if !q.conn.IsClosed() {
		if err := q.channel.Cancel(q.tag, true); err != nil {
			slog.Error("messaging queue - channel cancel failed: " + err.Error())
		}
		for channel := range q.channelPool {
			if err := channel.Cancel(q.tag, true); err != nil {
				slog.Error("messaging queue - channel cancel failed: " + err.Error())
			}
		}
		close(q.channelPool)
		if err := q.conn.Close(); err != nil {
			slog.Error("messaging queue - connection close failed: " + err.Error())
		}
	}
}

func (q *QueueProvider) initConn() (*amqp.Connection, error) {
	addr := fmt.Sprintf("amqp://%s:%s@%s:%d%s", q.username, q.password, q.addr, q.port, q.vhost)
	if conn, err := amqp.Dial(addr); err != nil {
		return nil, err
	} else {
		return conn, err
	}
}

func (q *QueueProvider) initChannel() (*amqp.Channel, error) {
	var (
		channel *amqp.Channel
		err     error
	)
	if channel, err = q.conn.Channel(); err != nil {
		q.conn.Close()
		return nil, err
	}

	if err := channel.ExchangeDeclare(
		q.exchange,
		q.exchangeType,
		false, //durable
		q.autoDelete,
		false,
		false,
		nil,
	); err != nil {
		channel.Close()
		q.conn.Close()
		return nil, err
	}

	channel.Qos(q.qos, 0, false)

	if _, err = channel.QueueDeclare(
		q.queue,
		false,        //durable
		q.autoDelete, //delete when ack
		false,
		false,
		q.args,
	); err != nil {
		channel.Close()
		q.conn.Close()
		return nil, err
	}

	if err = channel.QueueBind(
		q.queue,
		q.routingKey,
		q.exchange,
		false,
		nil,
	); err != nil {
		channel.Close()
		q.conn.Close()
		return nil, err
	}
	return channel, nil
}

// initChannel initializes a new AMQP channel from the pool
func (q *QueueProvider) getChannel() (*amqp.Channel, error) {
	select {
	// case channel := <-q.channelPool:
	// 	// Reuse existing channel from the pool
	// 	return channel, nil
	default:
		// Create a new channel if the pool is empty
		return q.initChannel()
	}
}

// releaseChannel releases an AMQP channel back to the pool
func (q *QueueProvider) releaseChannel(channel *amqp.Channel) {
	select {
	case q.channelPool <- channel:
		// Return the channel to the pool
	default:
		// Pool is full, close the channel
		channel.Close()
	}
}

// Run 运行队列
func (q *QueueProvider) Run() error {
	var err error

	if q.conn, err = q.initConn(); err != nil {
		return err
	}

	if q.channel, err = q.initChannel(); err != nil {
		return err
	}

	if q.handler != nil {
		var delivery <-chan amqp.Delivery
		if delivery, err = q.channel.Consume(
			q.queue,
			q.tag,
			false, //auto-act
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   //args
		); err != nil {
			q.channel.Close()
			q.conn.Close()
			return err
		}

		go q.Handle(delivery)
	}

	q.connNotify = q.conn.NotifyClose(make(chan *amqp.Error))
	q.channelNotify = q.channel.NotifyClose(make(chan *amqp.Error))

	return err
}

// ReConnect 重新连接队列
func (q *QueueProvider) ReConnect() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("ReConnect error", "error", err)
		}
	}()
	for {
		select {
		case err := <-q.connNotify:
			if err != nil {
				slog.Error("messaging queue - conn notifyclose:" + err.Error())
			}
		case err := <-q.channelNotify:
			if err != nil {
				slog.Error("messaging queue - channel notifyclose: " + err.Error())
			}

		case <-q.quit:
			return
		}

		// backstop
		if !q.conn.IsClosed() {
			if err := q.channel.Cancel(q.tag, true); err != nil {
				slog.Error("messaging queue - channel cancel failed: " + err.Error())
			}

			if err := q.conn.Close(); err != nil {
				slog.Error("messaging queue - connection cancel failed: " + err.Error())
			}
		}

		for err := range q.channelNotify {
			slog.Error(err.Error())
		}

		for err := range q.connNotify {
			slog.Error(err.Error())
		}

	quit:

		for {
			select {
			case <-q.quit:
				return
			default:
				slog.Info("messaging queue - reconnect")

				if err := q.Run(); err != nil {
					slog.Error("messaging queue - failCheck: " + err.Error())
					time.Sleep(time.Second * 5)
					continue
				}
				break quit
			}
		}
	}
}

// Handle 消息处理
func (q *QueueProvider) HandleBatch(delivery <-chan amqp.Delivery) {
	if q.handler == nil {
		return
	}

	for d := range delivery {
		go func(delivery amqp.Delivery) {
			if err := q.handler(delivery.Body); err == nil {
				delivery.Ack(false)
			} else {
				delivery.Reject(true)
			}
		}(d)
	}
}

// Handle 消息处理
func (q *QueueProvider) Handle(delivery <-chan amqp.Delivery) {
	if q.handler == nil {
		return
	}
	for d := range delivery {
		go func(delivery amqp.Delivery) {
			if err := q.handler(delivery.Body); err == nil {
				delivery.Ack(false)
			} else {
				delivery.Reject(true)
			}
		}(d)
	}
}

// PublishTo 发布到某个路由的Q里
func (q *QueueProvider) PublishTo(route string, msg []byte) error {
	//fmt.Printf("Publish to %s n \n ", route)
	if q == nil || q.channel == nil {
		err := fmt.Errorf("no channel valid %s", route)
		slog.Error(string(msg), "error", err)
		return err
	}

	return q.channel.Publish(
		q.exchange,
		route,
		false,
		false,
		amqp.Publishing{
			//ContentType: "application/json",
			Body: msg,
		},
	)
}

// Publish 发布一条消息
func (q *QueueProvider) Publish(msg []byte) error {
	return q.PublishTo(q.routingKey, msg)
}

// SafePublish 发送时独占channel 且会重试
func (q *QueueProvider) SafePublish(msg []byte) error {
	var err error
	for i := 0; i < len(q.channelPool)+1; i++ {
		err = q.quickPublish(msg)
		if err == nil {
			break
		}
	}
	if err != nil {
		slog.Error("QuickPublish", "error", err)
		return err
	}
	return nil
}

func (q *QueueProvider) quickPublish(msg []byte) error {
	channel, err := q.getChannel()
	if err != nil {
		slog.Error("QuickPublish getChannel", "error", err)
		return err
	}
	defer q.releaseChannel(channel)
	err = channel.Publish(
		q.exchange,
		q.routingKey,
		false,
		false,
		amqp.Publishing{
			//ContentType: "application/json",
			Body: msg,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// PublishToExchange
// agent跨越exchange进行消息投递
// 复用作为consumer的agent queue
// producer则无需声明自己的queue
func (q *QueueProvider) PublishToExchange(exchange, route string, msg []byte) error {
	return q.channel.Publish(
		exchange,
		route,
		false,
		false,
		amqp.Publishing{
			//ContentType: "application/json",
			Body: msg,
		},
	)
}

func (q *QueueProvider) Queue() string {
	return q.queue
}

func defaultHandler(msg []byte) error {
	fmt.Println(string(msg))
	return nil
}

// DefaultQueueProvider 默认队列
var DefaultQueueProvider *QueueProvider

func init() {
	DefaultQueueProvider = NewQueueProvider("onedotnet.asynctask", ExchangeDirect, "default", "default", false, defaultHandler)
	DefaultQueueProvider.Start()
}
