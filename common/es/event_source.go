/*
 * Copyright (c) 2023 fjw
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package es

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"runtime"
	"strings"
	"system-conf/common"
	"system-conf/common/log"
	"time"
)

type IEventSourceMessage interface {
	ToRaw() []byte
	ToJson() []byte
	ToBase64() []byte
}
type EventSourceResult struct {
	Code    int           `json:"code"`
	Message string        `json:"msg"`
	Cost    time.Duration `json:"cost,omitempty"`
	Err     error         `json:"err,omitempty"`
	Data    any           `json:"data,omitempty"`
}
type MessageBody struct {
	Id                 string `json:"id"`
	Src                string `json:"src"`
	Data               any    `json:"data,omitempty"`
	*EventSourceResult `json:",inline"`
}

func (m *MessageBody) ToJson() []byte {
	buf, _ := json.Marshal(m)
	return buf
}
func (m *MessageBody) ToRaw() []byte {
	switch data := m.Data.(type) {
	case []byte:
		return data
	case string:
		return []byte(data)
	default:
		buf, _ := json.Marshal(data)
		return buf
	}
}

func (m *MessageBody) ToBase64() []byte {
	buf, _ := json.Marshal(m.Data)
	if len(buf) > 1 {
		return buf[1 : len(buf)-1]
	}
	return nil
}

type MessageStrBody struct {
	Id   string `json:"id"`
	Src  string `json:"src"`
	Data string `json:"data"`
}

func (m *MessageStrBody) ToJson() []byte {
	buf, _ := json.Marshal(m)
	return buf
}
func (m *MessageStrBody) ToRaw() []byte {
	return []byte(m.Data)
}
func (m *MessageStrBody) ToBase64() []byte {
	buf, _ := json.Marshal(m.Data)
	if len(buf) > 1 {
		return buf[1 : len(buf)-1]
	}
	return nil
}

func (m *MessageBody) ToRawJson() []byte {
	buf, _ := json.Marshal(m)
	return buf
}

type MessageChan chan IEventSourceMessage

type EventSourceBroker struct {
	Id string
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan IEventSourceMessage

	dumpMap map[string]bool

	Logs []string

	// New client connections
	newClients chan MessageChan

	// Closed client connections
	closingClients chan MessageChan

	// Client connections registry
	clients map[MessageChan]bool

	closeSig chan struct{}
}

func (broker *EventSourceBroker) Close() {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("close sig chan failed:%v", e)
		}

	}()
	close(broker.closeSig)
}

func (broker *EventSourceBroker) SetDumpPile(name string, val bool) {
	broker.dumpMap[name] = val
}
func (broker *EventSourceBroker) ResetCloseSig() {
	broker.closeSig = make(chan struct{})
}

// Listen on different channels and act accordingly
func (broker *EventSourceBroker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:
			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan, _ := range broker.clients {
				clientMessageChan <- event
			}
		}
	}

}
func (broker *EventSourceBroker) PushJson(src string, data interface{}) (err error) {
	if buf, e := json.Marshal(data); e != nil {
		err = fmt.Errorf("failed to marshal json:%v", e)
		return
	} else {
		broker.Notifier <- &MessageBody{
			Id:   broker.Id,
			Src:  src,
			Data: buf,
		}
	}
	return
}
func (broker *EventSourceBroker) PushData(src string, data interface{}) (err error) {
	broker.Notifier <- &MessageBody{
		Id:   broker.Id,
		Src:  src,
		Data: data,
	}
	return
}
func (broker *EventSourceBroker) PushResult(result EventSourceResult) (err error) {
	broker.Notifier <- &MessageBody{
		Id:                broker.Id,
		Src:               "result",
		EventSourceResult: &result,
	}
	return
}

type EsBase64 struct {
	Buf []byte `json:"buf"`
}

func (m *EsBase64) Marshal() []byte {
	buf, _ := json.Marshal(m.Buf)
	if len(buf) > 1 {
		return buf[1 : len(buf)-1]
	}
	return nil
}

type EsJson struct {
	Data string `json:"data"`
}

func (m *EsJson) Marshal() []byte {
	buf, _ := json.Marshal(m)
	return buf
}

var esPrefix = []byte("data: ")
var esSuffix = []byte("\n\n")

const esDone = "------------ HTTP EVENT SOURCE DONE ------------"

func WriteEsData(rw http.ResponseWriter, raw bool, format string, msg IEventSourceMessage) {
	// Make sure that the writer supports flushing.
	if raw {
		// Raw JSON events, one per line
		rw.Write(msg.ToRaw())
	} else {
		rw.Write(esPrefix)
		switch format {
		case "base64":
			rw.Write(msg.ToBase64())
		case "json":
			rw.Write(msg.ToJson())
		default:
			rw.Write(msg.ToRaw())
			//rw.Write([]byte(" \b"))
			//fmt.Printf("msg:%s", msg)
		}
		rw.Write(esSuffix)
	}
}

// Implement the http.Handler interface.
// This allows us to wrap HTTP handlers (see auth_handler.go)
// http://golang.org/pkg/net/http/#Handler
func (broker *EventSourceBroker) ServeGin(c *gin.Context, msgs ...string) {

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	c.Writer.Header().Set("Content-Type", "text/event-stream;charset=UTF-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the EventSourceBroker's connections registry
	messageChan := make(MessageChan)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// "raw" query string option
	// If provided, send raw JSON lines instead of SSE-compliant strings.
	c.DefaultQuery("raw", "0")
	c.DefaultQuery("format", "raw")
	raw := false
	if v := common.ParseIntFromQuery(c, "raw"); v != nil && *v > 0 {
		raw = true
	}
	format := c.Query("format")

	// Listen to connection close and un-register messageChan

	notify := c.Writer.(http.CloseNotifier).CloseNotify()
	go func() {
	DONE:
		for {
			select {
			case <-c.Request.Context().Done():
				break DONE
			case <-notify:
				break DONE
			case <-broker.closeSig:
				break DONE
			}
		}
		broker.closingClients <- messageChan
		broker.PushData("EventSourceBroker", esDone)
		close(messageChan)

	}()
	for _, msg := range msgs {
		WriteEsData(c.Writer, raw, format, &MessageStrBody{
			Id:   broker.Id,
			Src:  "prefill",
			Data: msg,
		})
		flusher.Flush()
	}
	defer c.Status(http.StatusOK)
	// block waiting or messages broadcast on this connection's messageChan
	for {
		if msg, ok1 := <-messageChan; ok1 {
			WriteEsData(c.Writer, raw, format, msg)
			flusher.Flush()
		} else {
			break
		}
	}

}
func (broker *EventSourceBroker) DumpOutput(cx context.Context) {

	// Each connection registers its own message channel with the EventSourceBroker's connections registry
	messageChan := make(MessageChan)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// "raw" query string option
	// If provided, send raw JSON lines instead of SSE-compliant strings.

	// Listen to connection close and un-register messageChan

	go func() {
	DONE:
		for {
			select {
			case <-cx.Done():
				break DONE
			case <-broker.closeSig:
				break DONE
			}
		}
		broker.closingClients <- messageChan
		broker.PushData("EventSourceBroker", esDone)
		close(messageChan)

	}()

	// block waiting or messages broadcast on this connection's messageChan
	for {
		if msg, ok1 := <-messageChan; ok1 {
			log.Println(msg)
		} else {
			break
		}
	}
}
func (broker *EventSourceBroker) BindInput(name string, reader io.Reader) {
	buf := make([]byte, 4096)
	rd := bufio.NewReader(reader)
	for {
		var charSet common.Charset
		switch runtime.GOOS {
		case "windows":
			charSet = common.GB18030
		default:
			charSet = common.UTF8
		}

		//if line, e := rd.ReadString('\n'); e != nil {
		//	if e == io.EOF {
		//		break
		//	}
		//	log.Println("Error reading from pipe:", e)
		//	break
		//} else {
		//	msg := common.ConvertByte2String([]byte(line), charSet)
		//	if name != "stdout" {
		//		log.Warnf("data from %s: %s", name, msg)
		//	}
		//	broker.Notifier <- &MessageBody{
		//		Id:   broker.Id,
		//		Src:  name,
		//		Data: msg,
		//	}
		//}
		if n, e := rd.Read(buf); e == nil {
			if n > 0 {
				msg := common.ConvertByte2String(buf[:n], charSet)
				logSz := len(broker.Logs)
				if msg == "\r" {
					if logSz > 0 {
						broker.Logs[logSz-1] += "\r"
					}
				} else {
					if logSz > 0 && strings.HasSuffix(broker.Logs[logSz-1], "\r") {
						broker.Logs[logSz-1] = msg
					} else {
						broker.Logs = append(broker.Logs, msg)
					}
				}
				//log.Printf("got log(%d):%s \n", len(msg), msg)
				//if strings.HasPrefix(msg, "frame=") && len(broker.Logs) > 0 {
				//	// 如果消息以 "frame=" 开头且 broker.Logs 不为空，替换最后一个元素
				//	broker.Logs[len(broker.Logs)-1] = msg
				//} else {
				//	// 否则，直接追加消息
				//	broker.Logs = append(broker.Logs, msg)
				//}
				if v, ok := broker.dumpMap[name]; ok && v {
					log.Warnf("data from %s: %s", name, msg)
				}
				//if name != "stdout" {
				//	log.Warnf("data from %s: %s", name, msg)
				//}
				broker.Notifier <- &MessageBody{
					Id:   broker.Id,
					Src:  name,
					Data: msg,
				}
			}
		} else {
			if e == io.EOF {
				break
			}
			log.Println("Error reading from pipe:", e)
			break
		}
	}
}
func NewEventStreamBroker() (broker *EventSourceBroker) {
	// Instantiate a broker
	broker = &EventSourceBroker{
		Id:             uuid.New().String(),
		Notifier:       make(chan IEventSourceMessage, 1024),
		dumpMap:        make(map[string]bool),
		newClients:     make(chan MessageChan),
		closingClients: make(chan MessageChan),
		clients:        make(map[MessageChan]bool),
		closeSig:       make(chan struct{}),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}

func UseEventSource(in io.Reader) func(c *gin.Context) {
	return func(c *gin.Context) {
		flusher := c.Writer.(http.Flusher)
		buf := make([]byte, 4096)
		c.Writer.WriteString("------ HTTP EVENT SOURCE STARTED -------\n")
		for {
			if n, e1 := in.Read(buf); e1 == nil && n > 0 {
				c.Writer.Write(buf[:n])
				flusher.Flush()
			} else if e1 != nil {
				log.Printf("session read failed:%v", e1)
				c.Writer.WriteString("------ HTTP EVENT SOURCE FINISHED -------\n")
				break
			}
		}
	}
}
