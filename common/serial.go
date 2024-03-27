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

package common

import (
	"context"
	"go.bug.st/serial"
	"system-conf/common/log"
	"time"
)

type SerialParams struct {
	Name        string `json:"name"`
	serial.Mode `json:",inline"`
	Enabled     bool `json:"enabled"`
}
type SerialCtx struct {
	Params  *SerialParams
	Port    serial.Port
	runFlag bool
	cc      context.Context
	cl      context.CancelFunc
}

func NewSerialCtx(params *SerialParams) (ctx *SerialCtx) {
	return &SerialCtx{Params: params}
}

func (ctx *SerialCtx) open() (err error) {
	ctx.Port, err = serial.Open(ctx.Params.Name, &ctx.Params.Mode)
	return
}
func (ctx *SerialCtx) close() {
	if ctx.Port != nil {
		ctx.Port.Close()
	}
}
func (ctx *SerialCtx) RunAsSlave(cb func(data []byte) []byte) {
	if !ctx.runFlag {
		ctx.cc, ctx.cl = context.WithCancel(context.Background())
		ctx.runFlag = true
		go ctx.run(cb)
	}
}
func (ctx *SerialCtx) Stop() {
	if ctx.runFlag {
		ctx.runFlag = false
		ctx.close()
		if ctx.cc != nil {
			<-ctx.cc.Done()
		}
	}
}

func (ctx *SerialCtx) run(cb func(data []byte) []byte) {
	log.Printf("serial ctx goroutine started")
	buf := make([]byte, 128)
	var err error
	for ctx.runFlag {
		if ctx.Port == nil {
			if err = ctx.open(); err != nil {
				log.Warnf("failed to open serial:%v", err)
				time.Sleep(time.Second * 5)
			}
			continue
		} else {
			var sz int
			sz, err = ctx.Port.Read(buf)
			if err != nil {
				ctx.close()
			}
			if sz > 0 && cb != nil {
				recv := cb(buf[:sz])
				if len(recv) > 0 {
					ctx.Port.Write(recv)
				}
			}
		}

	}
	ctx.cl()
	log.Printf("serial ctx goroutine exited")

}
