package mqtt

import (
	"context"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"system-conf/common/log"
	"time"
)

func NewSession(opt *Options) (session *Session, err error) {
	if opt.Addr == "" {
		err = fmt.Errorf("invalid value of addr")
		return
	}
	if opt.Un == "hjjn" && opt.Pw == "" {
		opt.Pw = "public"
	}
	if opt.Un == "bymqtt" && opt.Pw == "" {
		opt.Pw = "bymqtt.public"
	}

	s := &Session{Options: opt, Subscriptions: make(map[string]*SubscribeObject)}
	opts := MQTT.NewClientOptions()
	//opts.CredentialsProvider = MQTT.tok
	addrs := strings.Split(opt.Addr, ",")
	if len(addrs) == 0 {
		err = fmt.Errorf("addr list is empty")
		return
	}

	for _, addr := range addrs {
		opts.AddBroker(addr)
	}
	switch opt.StorePath {
	case "":
	case "@mem":
		s.MemStore = MQTT.NewMemoryStore()
		opts.SetStore(s.MemStore)
	default:
		if !filepath.IsAbs(opt.StorePath) {
			tmp := opt.StorePath
			if opt.BasePath != "" {
				tmp = filepath.Join(opt.BasePath, opt.StorePath)
			}
			if v, e := filepath.Abs(tmp); e == nil {
				opt.StorePath = v
			}
		}
		fmt.Printf("storage path: %s\n", opt.StorePath)
		s.FileStroe = MQTT.NewFileStore(opt.StorePath)
		opts.SetStore(s.FileStroe)
	}

	if opt.ClientId != "" {
		opts.SetClientID(opt.ClientId)
	} else {
		hostname, _ := os.Hostname()
		clientId := fmt.Sprintf("%s_%d@%s_%s", log.ProcName, os.Getpid(), hostname, uuid.New().String())
		opts.SetClientID(clientId)
	}
	if opt.Un != "" {
		opts.SetUsername(opt.Un)
	}
	if opt.Pw != "" {
		opts.SetPassword(opt.Pw)
	}
	opts.SetConnectTimeout(opt.Timeout)
	opts.SetAutoReconnect(true)
	opts.SetResumeSubs(true)
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		reader := client.OptionsReader()
		if opt.Debug > 0 {
			fmt.Printf("mqtt client is connecting. client id: %v\n", reader.ClientID())
		}
		if v, ok := SessionCache.Load(client); ok {
			if s, ok := v.(*Session); ok {
				s.onReconnect(client)
			}
		}
	})

	opts.SetDefaultPublishHandler(session.DefaultHandler)
	s.Client = MQTT.NewClient(opts)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	err = s.connectV1()
	if err == nil {
		session = s
		SessionCache.Store(s.Client, session)
	}
	return
}
func NewPersistSession(opt *Options) (session *Session) {
	s := &Session{Options: opt, Subscriptions: make(map[string]*SubscribeObject)}
	opts := MQTT.NewClientOptions()
	if opt.Addr == "" {
		log.Panic("invalid value of addr")
	}
	addrs := strings.Split(opt.Addr, ",")
	if len(addrs) == 0 {
		log.Panic("addr list is empty")
	}
	for _, addr := range addrs {
		opts.AddBroker(addr)
	}
	if opt.Un == "hjjn" && opt.Pw == "" {
		opt.Pw = "public"
	}
	if opt.Un == "bymqtt" && opt.Pw == "" {
		opt.Pw = "bymqtt.public"
	}

	//opts.SetProtocolVersion(4)
	if opt.ClientId != "" {
		opts.SetClientID(opt.ClientId)
	} else {
		hostname, _ := os.Hostname()
		clientId := fmt.Sprintf("%s_%d@%s#%s", log.ProcName, os.Getpid(), hostname, log.Sn)
		opts.SetClientID(clientId)
	}
	if opt.Un != "" {
		opts.SetUsername(opt.Un)
	}
	if opt.Pw != "" {
		opts.SetPassword(opt.Pw)
	}
	opts.SetCredentialsProvider(func() (un, pw string) {
		return opt.Un, opt.Pw
	})

	opts.SetProtocolVersion(5)
	opts.SetConnectRetry(true)
	if opt.RetryInterval > 0 {
		opts.SetConnectRetryInterval(opt.RetryInterval)
	} else {
		opts.SetConnectRetryInterval(time.Second * 20)
	}
	//if len(opt.StorePath) == 0 {
	//	opt.StorePath = "./mqtt/" + opt.ClientId
	//}
	switch opt.StorePath {
	case "":
	case "@mem":
		s.MemStore = MQTT.NewMemoryStore()
		opts.SetStore(s.MemStore)
	default:

		if !filepath.IsAbs(opt.StorePath) {
			tmp := opt.StorePath
			if opt.BasePath != "" {
				tmp = filepath.Join(opt.BasePath, opt.StorePath)
			}
			if v, e := filepath.Abs(tmp); e == nil {
				opt.StorePath = v
			}
		}
		fmt.Printf("storage path: %s\n", opt.StorePath)
		s.FileStroe = MQTT.NewFileStore(opt.StorePath)
		opts.SetStore(s.FileStroe)
	}

	opts.SetConnectTimeout(opt.Timeout)
	opts.SetAutoReconnect(true)
	//opts.SetResumeSubs(true)
	opts.SetReconnectingHandler(func(client MQTT.Client, opts *MQTT.ClientOptions) {
		log.Printf("try to reconnect mqtt. ")
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		times := atomic.AddInt64(&s.ReconnectCount, 1)
		reader := client.OptionsReader()
		cid := reader.ClientID()
		defer log.Warnf("mqtt client(%v) is connected(connect times:%d).", cid, times)
		if times > 1 {
			if v, ok := SessionCache.Load(cid); ok {
				if ss, ok1 := v.(*Session); ok1 {
					ss.Client = client
					func(sess *Session) {
						log.Println("resubscribe after 5 seconds")
						time.Sleep(time.Second * 5)
						sess.Resubscribe()
					}(ss)
				}
				log.Warnf("resubscribed ok\n")
			}
		}

	})

	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		log.Warnf("mqtt connection lost: %v", err)
	})
	opts.SetDefaultPublishHandler(session.DefaultHandler)

	s.Client = MQTT.NewClient(opts)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	if e := s.connectV1(); e != nil {
		log.Printf("failed to connect mqtt broker; err:%v", e)
	}
	defer SessionCache.Store(opts.ClientID, s)
	session = s

	return
}
func NewSessionEx(opt *Options) (session *Session, err error) {
	if opt.Addr == "" {
		err = fmt.Errorf("invalid value of addr")
		return
	}
	if opt.Un == "hjjn" && opt.Pw == "" {
		opt.Pw = "public"
	}
	if opt.Un == "bymqtt" && opt.Pw == "" {
		opt.Pw = "bymqtt.public"
	}

	s := &Session{Options: opt, Subscriptions: make(map[string]*SubscribeObject)}
	opts := MQTT.NewClientOptions()
	//opts.CredentialsProvider = MQTT.tok
	addrs := strings.Split(opt.Addr, ",")
	if len(addrs) == 0 {
		err = fmt.Errorf("addr list is empty")
		return
	}

	for _, addr := range addrs {
		opts.AddBroker(addr)
	}
	switch opt.StorePath {
	case "":
	case "@mem":
		s.MemStore = MQTT.NewMemoryStore()
		opts.SetStore(s.MemStore)
	default:
		if !filepath.IsAbs(opt.StorePath) {
			tmp := opt.StorePath
			if opt.BasePath != "" {
				tmp = filepath.Join(opt.BasePath, opt.StorePath)
			}
			if v, e := filepath.Abs(tmp); e == nil {
				opt.StorePath = v
			}
		}
		fmt.Printf("storage path: %s\n", opt.StorePath)
		s.FileStroe = MQTT.NewFileStore(opt.StorePath)
		opts.SetStore(s.FileStroe)
	}

	if opt.ClientId != "" {
		opts.SetClientID(opt.ClientId)
	} else {
		hostname, _ := os.Hostname()
		clientId := fmt.Sprintf("%s_%d@%s_%s", log.ProcName, os.Getpid(), hostname, uuid.New().String())
		opts.SetClientID(clientId)
	}
	if opt.Un != "" {
		opts.SetUsername(opt.Un)
	}
	if opt.Pw != "" {
		opts.SetPassword(opt.Pw)
	}
	opts.SetConnectTimeout(opt.Timeout)
	opts.SetAutoReconnect(true)
	opts.SetResumeSubs(true)
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		reader := client.OptionsReader()
		if opt.Debug > 0 {
			fmt.Printf("mqtt client is connecting. client id: %v\n", reader.ClientID())
		}
		if v, ok := SessionCache.Load(client); ok {
			if s, ok := v.(*Session); ok {
				s.onReconnect(client)
			}
		}
	})

	opts.SetDefaultPublishHandler(session.DefaultHandler)
	s.Client = MQTT.NewClient(opts)
	SessionCache.Store(s.Client, s)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	err = s.connectV1()
	if err == nil {
		session = s
	}
	return
}
func NewMqttSession(addr, cid, un, pw string, debug int, timeout time.Duration) (ms *Session, err error) {
	var opt *Options
	opt = &Options{
		Addr:     addr,
		ClientId: cid,
		Un:       un,
		Pw:       pw,
		Timeout:  timeout,
		Debug:    debug,
	}
	if opt.Un == "hjjn" && opt.Pw == "" {
		opt.Pw = "hjjn"
	}
	ms, err = NewSession(opt)
	return
}
