package service

import (
	"os"
	"syscall"
)

// 在大明老师基础课 copy 的
var (
	ShutdownSignals = []os.Signal{
		os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGSTOP,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP,
		syscall.SIGABRT, syscall.SIGSYS, syscall.SIGTERM,
	}
)
