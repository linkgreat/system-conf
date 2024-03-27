package mqtt

import (
	"context"
	"fmt"
	"system-conf/common/log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"sync"
)

//	type Options struct {
//		Addr     string        `yaml:"addr" json:"addr"`
//		ClientId string        `yaml:"cid,omitempty" json:"cid,omitempty"`
//		Un       string        `yaml:"un,omitempty" json:"un,omitempty"`
//		Pw       string        `yaml:"pw,omitempty" json:"pw,omitempty"`
//		MemStore    bool          `yaml:"store" json:"store"`
//		Timeout  time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`
//		Debug    int           `yaml:"debug" json:"debug"`
//	}
type Session struct {
	*Options
	TopicRoot      string
	Client         MQTT.Client
	MemStore       *MQTT.MemoryStore
	FileStroe      *MQTT.FileStore
	Closed         bool
	ctx            context.Context
	cancel         context.CancelFunc
	Subscriptions  map[string]*SubscribeObject
	ReconnectCount int64
}

type MessageHandler func(msg MQTT.Message)

type SubscribeObject struct {
	Topic   string
	Qos     byte
	Handler MessageHandler
}

func NewSubscribeObject(topic string, qos int, handler MessageHandler) (so *SubscribeObject) {
	so = &SubscribeObject{
		Topic:   topic,
		Qos:     byte(qos),
		Handler: handler,
	}
	return so
}

var SessionCache = sync.Map{}

func (m *Session) Resubscribe() {
	for _, v := range m.Subscriptions {
		log.Printf("try to resubscribe: %s", v.Topic)
		if e := m.subscribe(v); e != nil {
			log.Warnf("failed to resubscribe: %s; err:%v", v.Topic, e)
		}
	}
}
func (m *Session) onReconnect(client MQTT.Client) {
	for _, v := range m.Subscriptions {
		_ = m.subscribe(v)
	}
}
func (m *Session) connect() (err error) {
	if m.Client != nil {
		token := m.Client.Connect()
		if !token.WaitTimeout(m.Timeout) {
			err = fmt.Errorf("connection timeout")
		} else {
			err = token.Error()
		}
	} else {
		err = fmt.Errorf("client is nil")
	}
	return
}
func (m *Session) connectV1() (err error) {
	if m.Client != nil {
		token := m.Client.Connect()
		token.Wait()
		err = token.Error()
	} else {
		err = fmt.Errorf("client is nil")
	}
	return
}
func (m *Session) DefaultHandler(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("got msg %v from %v: %v\n", msg.MessageID(), msg.Topic(), string(msg.Payload()))
	msg.Ack()
}
func (m *Session) subscribe(so *SubscribeObject) (err error) {
	token := m.Client.Subscribe(so.Topic, so.Qos, func(c MQTT.Client, msg MQTT.Message) {
		if so.Handler != nil {
			so.Handler(msg)
		}
	})
	bTimeout := false
	if m.Timeout != 0 {
		bTimeout = !token.WaitTimeout(m.Timeout)
	} else {
		token.Wait()
	}
	if bTimeout {
		err = fmt.Errorf("timeout")
	} else {
		err = token.Error()
	}
	return
}

func (m *Session) subscribeV1(so *SubscribeObject) (err error) {
	token := m.Client.Subscribe(so.Topic, so.Qos, func(c MQTT.Client, msg MQTT.Message) {
		if so.Handler != nil {
			so.Handler(msg)
		}
	})
	token.Wait()
	err = token.Error()
	return
}
func (m *Session) Subscribe(topic string, qos int, handler MessageHandler) (err error) {
	so := NewSubscribeObject(topic, qos, handler)
	err = m.subscribe(so)
	if err != nil {
		log.Warnf("subscribe %s failed; err:%v", topic, err)
	}
	m.Subscriptions[so.Topic] = so
	return
}
func (m *Session) SubscribeV1(topic string, qos int, handler MessageHandler) (err error) {
	so := NewSubscribeObject(topic, qos, handler)
	err = m.subscribeV1(so)
	if err != nil {
		log.Warnf("subscribe %s failed; err:%v", topic, err)
	}
	m.Subscriptions[so.Topic] = so
	return
}
func (m *Session) Unsubscribe(topics ...string) {
	for _, topic := range topics {
		delete(m.Subscriptions, topic)
	}
	m.Client.Unsubscribe(topics...)
}
func (m *Session) Publish(topic string, qos int, retained bool, payload interface{}) error {
	return m.PublishAsync(topic, qos, retained, payload)
}

// PublishAsync Deprecated
func (m *Session) PublishAsync(topic string, qos int, retained bool, payload interface{}) (err error) {
	if m.Client == nil {
		err = fmt.Errorf("mqtt client is nil")
		return
	}
	token := m.Client.Publish(topic, byte(qos), retained, payload)
	if m.Debug > 0 {
		go func() {
			if !token.WaitTimeout(m.Timeout) {
				log.Warnf("publish %s timeout", topic)
			} else {
				if e := token.Error(); e != nil {
					log.Warnf("failed to publish %s; err:%v", topic, e)
				}
			}
		}()
	}
	return
}
func (m *Session) PublishSync(topic string, qos int, retained bool, payload interface{}) (err error) {
	if m.Client != nil {
		token := m.Client.Publish(topic, byte(qos), retained, payload)
		//if m.Timeout != 0 {
		//	if !token.WaitTimeout(m.Timeout) {
		//		return fmt.Errorf("timeout")
		//	}
		//} else {
		//
		//}
		token.Wait()
		err = token.Error()
	}

	return
}
func (m *Session) Close() {
	m.Closed = true
	m.Client.Disconnect(250)
	//<-m.ctx.Done()
}
