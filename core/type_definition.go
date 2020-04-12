package core

import (
  "github.com/libp2p/go-libp2p-core/peerstore"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
)

type Mpi interface {
  Close() error
  CloseChan() chan bool
  ErrorChan() chan error
  Check() bool

  Add(string) error
  Del(string) error
  Get(uint64) error

  Host() ExtHost
  Store() Store
  Start(string, int, ...string) error
}

type ExtHost interface {
  CloseChan() chan bool
  ErrorChan() chan error
  Check() bool
  PeerstoreProtocol(protocol.ID) (peerstore.Peerstore, error)
  NewPeer(protocol.ID) (peer.ID, error)
  Listen(protocol.ID, string)
  SelfStream(...protocol.ID) (SelfStream, error)

  host.Host
}

type Store interface {
  Close() error
  CloseChan() chan bool
  ErrorChan() chan error
  Add(string)
  List() []string
  Has(string) bool
  Del(string) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type MasterComm interface {
  Connect(int, peer.ID, bool)
  CheckPeer(int) bool
  Reset(int)

  SlaveComm
}

type SlaveComm interface {
  Close() error
  CloseChan() chan bool
  ErrorChan() chan error
  Check() bool
  Interface() Interface
  Send(int, string)
  Get(int) string
  InitConnection(int, peer.ID) error
}

type Interface interface {
  Close() error
  CloseChan() chan bool
  ErrorChan() chan error
  Check() bool
  Message() chan Message
  Request() chan int
  Push(string) error
}

type SelfStream interface {
  Reverse() (SelfStream, error)

  network.Stream
}

type Message struct {
  To int
  Content string
}
