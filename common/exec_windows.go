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
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"system-conf/common/log"
)

func CloseCurrentProcessGroup() error {
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE|syscall.PROCESS_QUERY_INFORMATION, false, uint32(os.Getpid()))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(handle)
	err = syscall.TerminateProcess(handle, 0)
	if err != nil {
		return err
	}
	return nil
}
func (m *Exec) Kill() {
	if m.Cmd != nil && m.Cmd.Process != nil {
		killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(m.Cmd.Process.Pid))
		go killCmd.Run()
		//err := killCmd.Run()
		//if err != nil {
		//	log.Warnf("杀掉进程时出错: %v", err)
		//}
		//m.Cmd.Process.Kill()
		//m.Cmd.Process.Release()
	}
}

func (m *Exec) Run() error {

	m.Cmd = exec.Command(m.ExecName, m.Args...)

	if m.WorkDir != "" {
		m.Cmd.Dir = m.WorkDir
	}
	m.Cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	stdout, _ := m.Cmd.StdoutPipe()
	defer stdout.Close()
	stderr, _ := m.Cmd.StderrPipe()
	defer stderr.Close()
	if m.pipeCb == nil {
		go m.readPipe(stdout, false)
		go m.readPipe(stderr, true)
	} else {
		m.pipeCb("stdout", stdout)
		m.pipeCb("stderr", stderr)
	}

	err := m.Cmd.Start()
	if err != nil {
		fmt.Println(err)
		return err
	}
	log.Printf("proc started: %s %v", m.Cmd.Path, m.Cmd.Args)

	m.isRunning = true
	defer func() {
		m.isRunning = false
	}()

	err = m.Cmd.Wait()
	if err != nil {
		err = fmt.Errorf("failed to start proc. err: %v", err)
	}
	return err
}
