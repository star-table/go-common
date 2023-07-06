package mqtt

import (
	"errors"
	"sync/atomic"

	"github.com/star-table/go-common/go-common/utils/stack"

	"github.com/go-kratos/kratos/v2/log"
	emitter "github.com/star-table/emitter-go-client"
)

var (
	ErrNoConfig = errors.New("no config")
	ErrNoClient = errors.New("no avalible mqtt client")
)

type MqttClient interface {
	NativeClient() (*emitter.Client, error)
	OnError(handler emitter.ErrorHandler)
	GenerateKey(channel, permission string, ttl int) (string, error)
	Publish(key, channel string, payload interface{}, ttl int) error
}

type Config struct {
	Address        string
	Host           string
	Port           int
	SecretKey      string
	ClientPoolSize int
}

type mqttClient struct {
	Config
	clients []*emitter.Client
	next    uint32
	log     *log.Helper
}

func GetMqttClient(conf *Config, logger log.Logger) (MqttClient, error) {
	if conf == nil {
		return nil, ErrNoConfig
	}

	clientCount := conf.ClientPoolSize
	if clientCount <= 0 {
		clientCount = 10
	}

	mc := mqttClient{
		Config:  *conf,
		clients: make([]*emitter.Client, 0, clientCount),
		next:    0,
		log:     log.NewHelper(logger),
	}

	// init each client
	for i := 0; i < clientCount; i++ {
		c, err := emitter.Connect(conf.Address, nil)
		if err != nil || !c.IsConnected() {
			// release resources
			for _, cc := range mc.clients {
				cc.Disconnect(0)
			}
			return nil, err
		}
		mc.clients = append(mc.clients, c)
	}

	return &mc, nil
}

func (m *mqttClient) NativeClient() (*emitter.Client, error) {
	return m.getClient()
}

func (m *mqttClient) OnError(handler emitter.ErrorHandler) {
	for _, c := range m.clients {
		c.OnError(handler)
	}
}

func (m *mqttClient) GenerateKey(channel, permissions string, ttl int) (string, error) {
	client, err := m.getClient()
	if err != nil {
		log.Errorf("[MQTT] GenerateKey cannot get client: %v. channel: %v, permissions: %s, ttl: %d.",
			err, channel, permissions, ttl)
		return "", err
	}

	key, err := client.GenerateKey(m.Config.SecretKey, channel, permissions, ttl)
	if err != nil {
		log.Errorf("[MQTT] GenerateKey failed: %v. channel: %v, permissions: %s, ttl: %d.",
			err, channel, permissions, ttl)
		return "", err
	}
	return key, nil
}

func (m *mqttClient) Publish(key, channel string, payload interface{}, ttl int) error {
	client, err := m.getClient()
	if err != nil {
		log.Errorf("[MQTT] Publish cannot get client: %v. key: %v, channel: %v.",
			err, key, channel)
		return err
	}
	if ttl > 0 {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					m.log.Errorf("Capture panic：%s\n Stack:%s", r, stack.GetStack())
				}
			}()
			if err = client.Publish(key, channel, payload, emitter.WithAtLeastOnce(), emitter.WithTTL(ttl)); err != nil {
				log.Errorf("[MQTT] Publish faild: %v. key: %v, channel: %v, payload: %v.",
					err, key, channel, payload)
			}
		}()
	} else {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					m.log.Errorf("Capture panic：%s\n Stack:%s", r, stack.GetStack())
				}
			}()
			if err = client.Publish(key, channel, payload, emitter.WithAtLeastOnce()); err != nil {
				log.Errorf("[MQTT] Publish faild: %v. key: %v, channel: %v, payload: %v.",
					err, key, channel, payload)
			}
		}()
	}
	return nil
}

// RoundRobin get next index
func (m *mqttClient) nextIndex() int {
	n := atomic.AddUint32(&m.next, 1)
	return (int(n) - 1) % len(m.clients)
}

func (m *mqttClient) getClient() (*emitter.Client, error) {
	if len(m.clients) == 0 {
		return nil, ErrNoClient
	}

	var client *emitter.Client
	var i int
	tryTimes := 0

RETRY:
	tryTimes += 1
	if tryTimes > len(m.clients) {
		return nil, ErrNoClient
	}

	i = m.nextIndex()
	client = m.clients[i]

	// 理论上不可能走进的分支
	if client == nil || !client.IsConnected() {
		m.log.Errorf("[MQTT] client [%v] is null or disconnected!", i)
		goto RETRY
	}

	// 默认自动重连是开启的，如果IsConnectionOpen=False则意味着底层连接正在自动重新连接状态
	// 此时跳过此连接，直接使用下一个连接
	if !client.IsConnectionOpen() {
		m.log.Infof("[MQTT] client [%v] conn is not open, will try next.", i)
		goto RETRY
	}

	return client, nil
}
