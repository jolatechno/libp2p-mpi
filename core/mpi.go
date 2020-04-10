package core

import (
  "errors"
  "context"
  "fmt"
  "bufio"
  "time"
  "strings"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/network"

  maddr "github.com/multiformats/go-multiaddr"
)

const (
  WaitDuration = time.Second
)

type addrList []maddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := maddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}

type Config struct {
  Url string
  Path string
  Ipfs_store string
  Maxsize uint64
  Base string
  BootstrapPeers addrList
}

func NewMpi(ctx context.Context, config Config) (Mpi, error) {
  host, err := NewHost(ctx, config.BootstrapPeers...)
  if err != nil {
    return nil, err
  }

  store, err := NewStore(config.Url, config.Path, config.Ipfs_store)
  if err != nil {
    return nil, err
  }

  proto := protocol.ID(config.Ipfs_store + config.Base)
  mpi := BasicMpi {
    Ctx:ctx,
    Pid: proto,
    Ended: false,
    Maxsize: config.Maxsize,
    Path: config.Path,
    EndChan: make(chan bool),
    Ipfs_store: config.Ipfs_store,
    MpiHost: host,
    MpiStore: store,
    MasterComms: make(map[int]MasterComm),
    SlaveComms: make(map[string]SlaveComm),
    Id: 0,
  }

  for _, f := range store.List() {
    mpi.Add(f)
  }

  go func() {
    for mpi.Check() {
      occupied, err := store.Occupied()
      if err != nil {
        return
      }

      left := config.Maxsize - occupied
      if left <= 0 {
        return
      }

      f, err := store.Get(left)
      if err != nil {
        return
      }

      err = mpi.Add(f)
      if err != nil {
        return
      }
    }
  }()

  go func() {
    <- store.CloseChan()
    if mpi.Check() {
      mpi.Close()
    }
  }()

  go func() {
    <- host.CloseChan()
    if mpi.Check() {
      mpi.Close()
    }
  }()

  return &mpi, nil
}

type BasicMpi struct {
  Ctx context.Context
  Pid protocol.ID
  Ended bool
  Maxsize uint64
  Path string
  EndChan chan bool
  Ipfs_store string
  MpiHost ExtHost
  MpiStore Store
  MasterComms map[int]MasterComm
  SlaveComms map[string]SlaveComm
  Id int
}

func (m *BasicMpi)Close() error {
  m.EndChan <- true
  m.Ended = true
  err := m.Store().Close()
  if err != nil {
    return err
  }

  err = m.Host().Close()
  if err != nil {
    return err
  }

  for _, comm := range m.SlaveComms {
    err = comm.Close()
    if err != nil {
      return err
    }
  }

  for _, comm := range m.MasterComms {
    err = comm.Close()
    if err != nil {
      return err
    }
  }

  return nil
}

func (m *BasicMpi)CloseChan() chan bool {
  return m.EndChan
}

func (m *BasicMpi)Check() bool {
  return !m.Ended
}

func (m *BasicMpi)Add(f string) error {
  if !m.Store().Has(f) {
    err := m.Store().Dowload(f)
    if err != nil {
      return err
    }
  }

  proto := protocol.ID("/" + f + "/" + string(m.Pid))
  m.Host().Listen(proto, "/" + f + "/" + m.Ipfs_store)

  fmt.Println("Setting StreamHandler, proto : ", proto) //--------------------------

  m.Host().SetStreamHandler(proto, func(stream network.Stream) {

    fmt.Println("StreamHandler 0, proto : ", proto) //--------------------------

    rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
    str, err := rw.ReadString('\n')
    if err != nil {
      fmt.Println("StreamHandler 0, err : ", err) //--------------------------
      return
    }

    fmt.Println("StreamHandler 1") //--------------------------

    param, err := ParamFromString(str[:len(str) - 1])
    if err != nil {
      fmt.Println("StreamHandler 1, err : ", err) //--------------------------
      return
    }

    fmt.Println("StreamHandler 2") //--------------------------

    inter, err := NewInterface(f, param.N, param.Idx)
    if err != nil {
      fmt.Println("StreamHandler 2, err : ", err) //--------------------------
      return
    }

    fmt.Println("StreamHandler 3") //--------------------------

    comm, err := NewSlaveComm(m.Ctx, m.Host(), rw, proto, inter, param)
    if err != nil {
      fmt.Println("StreamHandler 3, err : ", err) //--------------------------
      return
    }

    m.SlaveComms[param.Id] = comm
    go func(id string){
      <- comm.CloseChan()
      delete(m.SlaveComms, id)
    }(param.Id)
  })
  return nil
}

func (m *BasicMpi)Del(f string) error {
  err := m.Store().Del(f)
  if err != nil {
    return err
  }

  proto := protocol.ID(f + string(m.Pid))
  m.Host().RemoveStreamHandler(proto)
  return nil
}

func (m *BasicMpi)Host() ExtHost {
  return m.MpiHost
}

func (m *BasicMpi)Store() Store {
  return m.MpiStore
}

func (m *BasicMpi)Get(maxsize uint64) error {
  f, err := m.MpiStore.Get(maxsize)
  if err != nil {
    return err
  }

  return m.Add(f)
}

func (m *BasicMpi)Start(file string, n int, args ...string) error {
  if !m.MpiStore.Has(file) {
    return errors.New("no such file")
  }

  fmt.Println("Start 0") //--------------------------

  inter, err := NewInterface(m.Path + file, n, 0, args...)
  if err != nil {
    fmt.Println("Start 0, err : ", err) //--------------------------
    return err
  }

  fmt.Println("Start 1") //--------------------------

  id := m.Id
  m.Id++

  proto := protocol.ID(fmt.Sprintf("/%s/%s", file, m.Pid))
  StringId := fmt.Sprintf("%d/%s", id, m.Host().ID())

  comm, err := NewMasterComm(m.Ctx, m.Host(), n, proto, inter, StringId)

  if err != nil {
    fmt.Println("Start 1, err : ", err) //--------------------------
    return err
  }

  m.MasterComms[id] = comm
  go func() {
    <- comm.CloseChan()
    delete(m.MasterComms, id)
  }()

  return nil
}
