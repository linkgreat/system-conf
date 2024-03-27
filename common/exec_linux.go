//go:build linux

package common

import (
	"fmt"
	exec "golang.org/x/sys/execabs"
	"sync"
	"syscall"
	"system-conf/common/log"
)

var ProcessMap = sync.Map{}

func CloseCurrentProcessGroup() error {
	gid := syscall.Getgid()
	pid := syscall.Getpid()
	log.Printf("got process group id: %d; pid:%d", gid, pid)
	ProcessMap.Range(func(key, value any) bool {
		defer ProcessMap.Delete(key)
		log.Warnf("try to term group process:%v", value)
		if p, ok := key.(int); ok {
			if e := syscall.Kill(-p, syscall.SIGTERM); e != nil {
				log.Warnf("failed to term group process; err:%v", e)
				_ = syscall.Kill(p, syscall.SIGTERM)
			}
		}
		if v, ok := value.(*exec.Cmd); ok {
			if proc := v.Process; proc != nil {
				_ = proc.Kill()
			}
		}
		return true
	})
	syscall.Exit(0)
	return nil
}
func (m *Exec) Kill() {
	if m.Cmd != nil && m.Cmd.Process != nil {
		pid := m.Cmd.Process.Pid
		log.Warnf("try to get pgid")
		if pgid, e := syscall.Getpgid(pid); e != nil {
			log.Warnf("failed to get pgid of %d", pid)
		} else {
			log.Warnf("try to term pid:%d pgid:%d", pid, pgid)
		}
		if pid != 0 {
			log.Warnf("try to kill process group of pid:%d", pid)
			if e := syscall.Kill(-pid, syscall.SIGTERM); e != nil {
				log.Warnf("failed to kill process group of pid:%d; err:%v", pid, e)
			}
		}
		_ = m.Cmd.Process.Kill()
		_ = m.Cmd.Process.Release()
	}
}

func (m *Exec) Run() error {
	m.Cmd = exec.Command(m.ExecName, m.Args...)
	if m.WorkDir != "" {
		m.Cmd.Dir = m.WorkDir
	}
	uid := syscall.Getuid()
	gid := syscall.Getgid()
	log.Warnf("exec cmd will be run by gid:%d; pid:%d; uid:%d", gid, syscall.Getpid(), uid)
	m.Cmd.SysProcAttr = &syscall.SysProcAttr{

		Credential: &syscall.Credential{
			Uid: uint32(syscall.Getuid()),
			Gid: uint32(gid),
		},
		Setpgid: true,
		Pgid:    0,
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
	pid := m.Cmd.Process.Pid
	log.Printf("proc(%d) started: %s %v", pid, m.Cmd.Path, m.Cmd.Args)
	log.Warnf("process %d started", pid)
	if pid > 0 && !m.ignoreParentExitTerSig {
		ProcessMap.Store(pid, m.Cmd)
		defer ProcessMap.Delete(pid)
	}

	m.runFlag = true
	defer func() {
		m.runFlag = false
	}()
	err = m.Cmd.Wait()

	if err != nil {
		err = fmt.Errorf("failed to start proc. err: %v", err)
	}
	return err
}
