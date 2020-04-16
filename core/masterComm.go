package core

import (
  "fmt"
  "context"
  //"time"
  "sync"
  "bufio"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-core/peer"
)

func NewMasterComm(ctx context.Context, host ExtHost, n int, base protocol.ID, id string, file string, args ...string) (_ MasterComm, err error) {
  inter, err := NewInterface(file, n, 0, args...)
  if err != nil {
    return nil, err
  }

  Addrs := make([]peer.ID, n)
  for i, _ := range Addrs {
    if i == 0 {
      Addrs[i] = host.ID()
    } else {
      Addrs[i], err = host.NewPeer(base)
      if err != nil {
        return nil, err
      }
    }
  }

  remotes := make([]Remote, n)
  comm := BasicMasterComm {
    Addrs: &Addrs,
    Comm: BasicSlaveComm {
      Ctx: ctx,
      Inter: inter,
      Id: id,
      N: n,
      Idx: 0,
      CommHost: host,
      Base: protocol.ID(fmt.Sprintf("%s/%s", id, string(base))),
      Pid: base,
      Remotes: &remotes,
      Standard: NewStandardInterface(),
    },
  }

  comm.Comm.SetErrorHandler(func(err error) {
    comm.Raise(err)
  })

  comm.Comm.SetCloseHandler(func() {
    if comm.Check() {
      comm.Close()
    }
  })

  state := 0

  var wg sync.WaitGroup

  wg.Add(n - 1)

  reseted := make([]bool, n)

  for j := 1; j < n; j++ {
    i := j

    (*comm.Comm.Remotes)[i], err = NewRemote()
    if err != nil {
      return nil, err
    }

    comm.SlaveComm().Remote(i).SetCloseHandler(func() {
      comm.Close()
    })

    go func(wp *sync.WaitGroup) {
      comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
        if state == 0 {
          reseted[i] = true
          wp.Done()
        }

        comm.Reset(i)

        comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
          comm.Reset(i)
        })
      })

      comm.Connect(i, (*comm.Addrs)[i], true)

      <- comm.SlaveComm().Remote(i).GetHandshake()

      if !reseted[i] {
        wp.Done()
      }
    }(&wg)
  }

  fmt.Println("[MasterComm] Handshake 0.5") //--------------------------

  wg.Wait()

  fmt.Println("[MasterComm] Handshake 1") //--------------------------

  state = 1

  N := 0
  for _, s := range reseted {
    if !s {
      N++
    }
  }

  var wg2 sync.WaitGroup

  wg2.Add(N - 1)

  for j := 1; j < n; j ++ {
    i := j

    if !reseted[i] {
      continue
    }

    comm.SlaveComm().Remote(i).SendHandshake()

    go func(wp *sync.WaitGroup) {
      comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
        if state == 1 {
          reseted[i] = true
          wp.Done()
        }

        comm.Reset(i)

        comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
          comm.Reset(i)
        })
      })

      <- comm.SlaveComm().Remote(i).GetHandshake()

      if !reseted[i] {
        wp.Done()
      }
    }(&wg2)
  }

  fmt.Println("[MasterComm] Handshake 1.5") //--------------------------

  wg2.Wait()

  fmt.Println("[MasterComm] Handshake 2") //--------------------------

  state = 2

  for j := 1; j < n; j++ {
    i := j

    comm.SlaveComm().Remote(i).SetErrorHandler(func(err error) {
      comm.Reset(i)
    })

    if !reseted[j] {
      comm.SlaveComm().Remote(j).SendHandshake()
    }
  }

  comm.SlaveComm().Start()

  return &comm, nil
}

type BasicMasterComm struct {
  Addrs *[]peer.ID
  Ctx context.Context
  Comm BasicSlaveComm
}

func (c *BasicMasterComm)Close() error {
  if c.SlaveComm().Check() {
    c.SlaveComm().Close()
  }

  return nil
}

func (c *BasicMasterComm)SetErrorHandler(handler func(error)) {
  c.SlaveComm().SetErrorHandler(handler)
}

func (c *BasicMasterComm)SetCloseHandler(handler func()) {
  c.SlaveComm().SetCloseHandler(handler)
}

func (c *BasicMasterComm)Raise(err error) {
  c.SlaveComm().Raise(err)
}

func (c *BasicMasterComm)Check() bool {
  return !c.SlaveComm().Check()
}

func (c *BasicMasterComm)SlaveComm() SlaveComm {
  return &c.Comm
}

func (c *BasicMasterComm)Connect(i int, addr peer.ID, init bool) {
  err := c.SlaveComm().Connect(i, addr)

  if err != nil {
    c.SlaveComm().Remote(i).Raise(err)
  } else {
    p := Param {
      Init: init,
      Idx: i,
      N: c.Comm.N,
      Id: c.Comm.Id,
      Addrs: c.Addrs,
    }

    writer := bufio.NewWriter(c.SlaveComm().Remote(i).Stream())

    _, err = writer.WriteString(fmt.Sprintf("%s\n", p.String()))
    if err != nil {
      c.SlaveComm().Remote(i).Raise(err)
      return
    }

    err = writer.Flush()
    if err != nil {
      c.SlaveComm().Remote(i).Raise(err)
      return
    }
  }
}

func (c *BasicMasterComm)Reset(i int) {

  fmt.Println("[MasterComm] reseting ", i) //--------------------------

  addr, err := c.SlaveComm().Host().NewPeer(c.Comm.Pid)
  if err != nil {
    c.Raise(err)
  }

  (*c.Addrs)[i] = addr
  c.Connect(i, addr, false)
}
