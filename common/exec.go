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
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type ProcessOutputCB func(isErr bool, line string)
type PipeCB func(string, io.ReadCloser)
type ProcessStatus struct {
	State  int    `json:"state"`
	ErrMsg string `json:"errMsg"`
}

const (
	ProcStatusBadArgs  = -1
	ProcStatusStopped  = 0
	ProcStatusStarting = 1
	ProcStatusStarted  = 2
	ProcStatusError    = 4
	ProcStatusDisabled = 8
)

var sysExitFlag chan bool

func init() {
	sysExitFlag = make(chan bool)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(sysExitFlag)
		time.Sleep(time.Millisecond * 100)
		syscall.Exit(-1)
	}()
}

type Exec struct {
	WorkDir                string
	ExecName               string
	Args                   []string
	cb                     ProcessOutputCB
	pipeCb                 PipeCB
	Cmd                    *exec.Cmd
	cx                     context.Context
	cl                     context.CancelFunc
	runFlag                bool
	isRunning              bool
	ignoreParentExitTerSig bool
	bCreateSession         bool
}

func NewExec(wd, execName string) (ex *Exec) {
	return &Exec{
		WorkDir:  wd,
		ExecName: execName,
		runFlag:  false,
	}
}

func (m *Exec) GetExitCode() int {
	if m.Cmd != nil && m.Cmd.ProcessState != nil {
		return m.Cmd.ProcessState.ExitCode()
	}
	return -1
}
func (m *Exec) SetIgnoreParentExitTermSig(val bool) *Exec {
	m.ignoreParentExitTerSig = val
	return m
}
func (m *Exec) SetPipeCallback(cb func(name string, reader io.ReadCloser)) *Exec {
	m.pipeCb = cb
	return m
}
func (m *Exec) SetCallback(cb ProcessOutputCB) {
	m.cb = cb
}
func (m *Exec) SetCreateSession(val bool) *Exec {
	m.bCreateSession = val
	return m
}

func (m *Exec) IsRunning() bool {
	return m.isRunning
}
func (m *Exec) StopRun() {
	m.Kill()
	if m.runFlag {
		m.runFlag = false
		<-m.cx.Done()
	}
}

func (m *Exec) StartRun(mode int, chStatus chan<- ProcessStatus) {
	if !m.runFlag {
		m.runFlag = true
		m.cx, m.cl = context.WithCancel(context.Background())
		if mode > 0 {
			go m.keepRun(mode, chStatus)
		}

	}
}
func (m *Exec) keepRun(mode int, chState chan<- ProcessStatus) {
	retry := 2
	for m.runFlag && retry > 0 {
		chState <- ProcessStatus{ProcStatusStarting, "启动中"}

		cx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(5 * time.Second)
			if m.isRunning {
				chState <- ProcessStatus{ProcStatusStarted, "已启动"}
			}
			ticker := time.NewTicker(time.Second * 10)
			defer ticker.Stop()
		DONE:
			for {
				select {
				case <-cx.Done():
					break DONE
				case <-ticker.C:
					chState <- ProcessStatus{ProcStatusStarted, "已启动"}
				}
			}
		}()
		go func() {
			<-sysExitFlag
			m.Kill()
		}()
		err := m.Run()
		cancel()

		if m.runFlag {
			if err != nil {
				if mode == 1 {
					retry--
				}
				chState <- ProcessStatus{ProcStatusError, err.Error()}
				fmt.Printf("process exited:%v and will be restarted in 3 seconds\n", err)
			} else {
				if mode == 1 {
					fmt.Printf("process exited without error\n")
					chState <- ProcessStatus{ProcStatusStopped, "done"}
					m.runFlag = false
				}
				fmt.Printf("process exited without error and will be restarted in 3 seconds\n")
				chState <- ProcessStatus{ProcStatusStopped, ""}

			}
			time.Sleep(time.Second * 5)
		}
	}
	m.cl()
	chState <- ProcessStatus{ProcStatusStopped, "finished"}
}
func (m *Exec) readPipe(stream io.ReadCloser, isErr bool) {
	reader := bufio.NewReader(stream)
	for {
		buf, err2 := reader.ReadBytes('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		line := ConvertByte2String(buf, UTF8)
		if m.cb != nil {
			m.cb(isErr, line)
		} else {
			fmt.Print(line)
		}
	}
}
