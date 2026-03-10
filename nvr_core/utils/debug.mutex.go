package utils

import (
    "log"
    "runtime"
    "sync"
)

type DebugMutex struct {
    sync.Mutex
}

func (m *DebugMutex) Lock() {
    _, file, line, _ := runtime.Caller(1)
    log.Printf("WAITING to lock at %s:%d\n", file, line)
    m.Mutex.Lock()
    log.Printf("LOCKED successfully at %s:%d\n", file, line)
}

func (m *DebugMutex) Unlock() {
    _, file, line, _ := runtime.Caller(1)
    log.Printf("UNLOCKING at %s:%d\n", file, line)
    m.Mutex.Unlock()
}