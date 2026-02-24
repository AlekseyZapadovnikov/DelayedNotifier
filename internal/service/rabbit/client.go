package rabbit

import (
	"log"
	"net"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
)

type RabbitClient struct {
	cfg *config.RabbitConfig
	mu  sync.Mutex

	ch              *amqp.Channel
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation

	conn            *amqp.Connection
	notifyConnClose chan *amqp.Error
	isReady         bool

	done chan struct{}
}

func NewRabbitClient(cfg *config.RabbitConfig) *RabbitClient {
	client := &RabbitClient{
		cfg:  cfg,
		done: make(chan struct{}),
	}

	go client.handleReconnect()

	return client
}

func (r *RabbitClient) Close() error {
	close(r.done)
	r.mu.Lock()
	defer r.mu.Unlock()

	r.isReady = false

	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// IsReady безопасная проверка статуса
func (r *RabbitClient) IsReady() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isReady
}

func (r *RabbitClient) connect() (*amqp.Connection, error) {
	amqpCfg := amqp.Config{
		Heartbeat: r.cfg.Client.Heartbeat,
		Properties: amqp.Table{
			"connection_name": r.cfg.Client.ServiceName,
		},
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, r.cfg.Client.ConnectTimeout)
		},
	}

	conn, err := amqp.DialConfig(r.cfg.Connection.ConnectionString(), amqpCfg)
	if err != nil {
		return nil, err
	}

	r.changeConnection(conn)
	return conn, nil
}

func (r *RabbitClient) changeConnection(connection *amqp.Connection) {
	r.conn = connection
	r.notifyConnClose = make(chan *amqp.Error, 1)
	r.conn.NotifyClose(r.notifyConnClose)
}

func (r *RabbitClient) changeChannel(channel *amqp.Channel) {
	r.ch = channel
	r.notifyChanClose = make(chan *amqp.Error, 1)
	r.notifyConfirm = make(chan amqp.Confirmation, 1)
	r.ch.NotifyClose(r.notifyChanClose)
	r.ch.NotifyPublish(r.notifyConfirm)
}

func (r *RabbitClient) handleReconnect() {
	for {
		r.mu.Lock()
		r.isReady = false
		r.mu.Unlock()

		log.Println("Attempting to connect to RabbitMQ...")
		conn, err := r.connect()
		if err != nil {
			log.Printf("Failed to connect: %v. Retrying in %v...", err, r.cfg.Client.ReinitTimeout)
			select {
			case <-r.done:
				return
			case <-time.After(r.cfg.Client.ReinitTimeout):
			}
			continue
		}

		if done := r.handleResetup(conn); done {
			break
		}
	}
}

func (r *RabbitClient) handleResetup(conn *amqp.Connection) bool {
	for {
		r.mu.Lock()
		r.isReady = false
		r.mu.Unlock()

		err := r.setupClient(conn)
		if err != nil {
			log.Printf("Failed to setup client: %v. Retrying in %v...", err, r.cfg.Client.ReinitTimeout)
			select {
			case <-r.done:
				return true
			case <-r.notifyConnClose:
				return false
			case <-time.After(r.cfg.Client.ReinitTimeout):
			}
			continue
		}

		log.Println("RabbitMQ connected and setup successfully")

		select {
		case <-r.done:
			return true
		case err := <-r.notifyConnClose:
			log.Printf("Connection closed: %v. Reconnecting...", err)
			return false
		case err := <-r.notifyChanClose:
			log.Printf("Channel closed: %v. Re-initializing...", err)
		}
	}
}