package core

import (
  "sync"
)

var (
  nilEndHandler = func() {
    return
  }
  nilErrorHandler = func(err error) {
    return
  }
)

func NewStandardInterface() standardFunctionsCloser {
  return &BasicFunctionsCloser {
    EndHandler: &nilEndHandler,
    ErrorHandler: &nilErrorHandler,
  }
}

type BasicFunctionsCloser struct {
  Mutex sync.Mutex
  Ended bool
  EndHandler *func()
  ErrorHandler *func(error)
}

func (b *BasicFunctionsCloser)Close() error {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  if !b.Ended {
    (*b.EndHandler)()

    b.Ended = true
  }

  return nil
}

func (b *BasicFunctionsCloser)Raise(err error) {
  if b.Check() {
    (*b.ErrorHandler)(err)
  }
}

func (b *BasicFunctionsCloser)Check() bool {
  b.Mutex.Lock()
  defer b.Mutex.Unlock()
  return !b.Ended
}

func (b *BasicFunctionsCloser)SetErrorHandler(handler func(error)) {
  b.ErrorHandler = &handler
}

func (b *BasicFunctionsCloser)SetCloseHandler(handler func()) {
  b.EndHandler = &handler
}
