package core

import (
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
)

type Mpi interface {
  Close() error
  Add(string) error
  CloseChan() chan bool
  Host() ExtHost
  Store() Store
  Get(uint64) error
  Start(string, int) error
}

type ExtHost interface {
  host.Host
  CloseChan() chan bool
  NewPeer(protocol.ID) peer.ID
}

type Store interface {
  Close() error
  CloseChan() chan bool
  Add(string)
  List() []string
  Has(string) bool
  Del(string) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type MasterComm interface {
  SlaveComm

  CheckPeer(int) bool
  Reset(int)
  Connect(int, peer.ID)
}

type SlaveComm interface {
  Close() error
  CloseChan() chan bool
  Check() bool
  Interface() Interface
  Send(int, string)
  Get(int) string
}

type Interface interface {
  Close() error
  CloseChan() chan bool
  Check() bool
  Message() chan Message
  Request() chan int
  Push(string) error
}

type Message struct {
  To int
  Content string
}
