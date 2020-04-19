package core

import (
  "io"
  "time"

  "github.com/libp2p/go-libp2p-core/peerstore"
  "github.com/libp2p/go-libp2p-core/peer"
  "github.com/libp2p/go-libp2p-core/host"
  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"
)

//-------

type standardFunctionsCloser interface {
  standardFunctions
  io.Closer
}

type standardFunctions interface {
  Check() bool
  Raise(error)
  SetCloseHandler(func())
  SetErrorHandler(func(error))
}

//-------

type Mpi interface {
  standardFunctionsCloser

  Add(string) error
  Del(string) error
  Get(uint64) error

  Host() ExtHost
  Store() Store
  Start(string, int, ...string) error
}

type ExtHost interface {
  host.Host
  standardFunctions

  PeerstoreProtocol(protocol.ID) (peerstore.Peerstore, error)
  NewPeer(protocol.ID) (peer.ID, error)
  Listen(protocol.ID, string)
  SelfStream(...protocol.ID) (SelfStream, error)
}

type Store interface {
  standardFunctionsCloser

  Add(string)
  List() []string
  Has(string) bool
  Del(name string, failed bool) error
  Dowload(string) error
  Occupied() (uint64, error)
  Get(uint64) (string, error)
}

type MasterComm interface {
  standardFunctionsCloser

  SlaveComm() SlaveComm
  Reset(idx int, ith_time int)
}

type SlaveComm interface {
  standardFunctionsCloser

  RequestReset(int)
  Start()
  Host() ExtHost
  Interface() Interface
  Remote(int) Remote
  Connect(int, peer.ID, ...string) error
}

type Remote interface {
  standardFunctionsCloser

  SetResetHandler(func(int, int))
  RequestReset(int, int)
  CloseRemote()
  SetPingInterval(time.Duration)
  SetPingTimeout(time.Duration)
  Stream() io.ReadWriteCloser
  Reset(io.ReadWriteCloser)
  Get() string
  GetHandshake() chan bool
  Send(string)
  SendHandshake()
}

type Interface interface {
  standardFunctionsCloser

  Start()
  SetResetHandler(func(int))
  SetMessageHandler(func(int, string))
  SetRequestHandler(func(int))
  Push(string) error
}

type SelfStream interface {
  Reverse() (SelfStream, error)

  network.Stream
}
