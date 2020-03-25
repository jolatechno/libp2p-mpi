package store

import (
  "context"
  "os"

  "github.com/jolatechno/mpi-peerstore/utils"
  "github.com/jolatechno/ipfs-mpi/core/ipfs-interface"
  "github.com/jolatechno/ipfs-mpi/core/api"

  "github.com/libp2p/go-libp2p-core/protocol"
  "github.com/libp2p/go-libp2p-discovery"
  "github.com/libp2p/go-libp2p-core/host"
  maddr "github.com/multiformats/go-multiaddr"
)

type Store struct {
  store map[file.File] Entry
  host host.Host
  routingDiscovery *discovery.RoutingDiscovery
  shell *file.IpfsShell
  api *api.Api
  protocol protocol.ID
  maxsize uint64
  path string
}

type Config struct {
	url string
	path string
	ipfs_store string
	BootstrapPeers []maddr.Multiaddr
	ListenAddresses []maddr.Multiaddr
	ProtocolID string
	maxsize uint64
	api_port int
	WriteTimeout int
	ReadTimeout int
}

func NewStore(ctx context.Context, host host.Host, config Config) (*Store, error) {
  store := make(map[file.File] Entry)
  proto := protocol.ID(config.ipfs_store + config.ProtocolID)

  routingDiscovery, err := utils.NewKadmeliaDHT(ctx, host, config.BootstrapPeers)
  if err != nil {
    return nil, err
  }

if _, err := os.Stat(config.path); os.IsNotExist(err) {
    os.MkdirAll(config.path, file.ModePerm)
  } else if err != nil {
    return nil, err
  }

  shell, err := file.NewShell(config.url, config.path, config.ipfs_store)
  if err != nil {
    return nil, err
  }

  api, err := api.NewApi(config.api_port, config.ReadTimeout, config.WriteTimeout)
  if err != nil {
    return nil, err
  }

  return &Store{ store:store, host:host, routingDiscovery:routingDiscovery, shell:shell, api:api, protocol:proto, maxsize:config.maxsize, path:config.path }, nil
}

func (s *Store)Init(ctx context.Context) error {
  files := (*s.shell).List()

  for _, f := range files {
    e := NewEntry(&s.host, s.routingDiscovery, f, s.shell, s.api, s.path)
    err := e.LoadEntry(ctx, s.protocol)
    if err != nil {
      return err
    }

    s.store[f] = *e
  }

  go func(){
    for{
      err := s.Get(ctx)
      if err != nil { //No new file to add
        return
      }
    }
  }()

  return nil
}

func (s *Store)Add(f file.File, ctx context.Context) error {
  e := NewEntry(&s.host, s.routingDiscovery, f, s.shell, s.api, s.path )

  err := e.InitEntry()
  if err != nil {
    return err
  }

  err = e.LoadEntry(ctx, s.protocol)
  if err != nil {
    return err
  }

  s.store[f] = *e
  return nil
}

func (s *Store)Del(f file.File) error {
  return s.shell.Del(f)
}

func (s *Store)Get(ctx context.Context) error {
  used, err := s.shell.Occupied()
  if err != nil {
    return err
  }

  f, err := s.shell.Get(s.maxsize - used)
  if err != nil {
    return err
  }

  return s.Add(*f, ctx)
}
