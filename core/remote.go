package core

import (
  "bufio"
  "fmt"

  "github.com/libp2p/go-libp2p-core/network"
)

func NewRemote(handshakeMessage int) (Remote, error) {
  return &BasicRemote {
    ReadChan: make(chan string),
    HandshakeChan: make(chan string),
    Sent: []string{},
    Rw: nil,
    Offset: 0,
    Received: 0,
    HandshakeMessage: handshakeMessage,
    ReceivedHandshakeMessage: 0,
    Standard: NewStandardInterface(),
  }, nil
}

type BasicRemote struct {
  ReadChan chan string
  HandshakeChan chan string
  Sent []string
  Rw *bufio.ReadWriter
  Offset int
  Received int
  HandshakeMessage int
  ReceivedHandshakeMessage int
  Standard BasicFunctionsCloser
}

func (r *BasicRemote)Send(msg string) {

  fmt.Printf("[Remote] Sending %q\n", msg) //--------------------------

  if r.ReceivedHandshakeMessage >= r.HandshakeMessage { //shouldn't be strictly greater
    r.Sent = append(r.Sent, msg)
  }

  fmt.Fprint(r.Rw, msg)
  r.Rw.Flush()
}

func (r *BasicRemote)Get() string {
  return <- r.ReadChan
}

func (r *BasicRemote)GetHandshake() string {
  return <- r.HandshakeChan
}

func (r *BasicRemote)Reset(stream *bufio.ReadWriter) {
  r.Rw = stream
  r.Offset = r.Received
  r.ReceivedHandshakeMessage = 0

  go func() {
    for r.Check() && r.Rw == stream {
      str, err := stream.ReadString('\n')
      if err != nil {
        r.Standard.Push(err)
        return
      }

      fmt.Printf("[Remote] Received %q\n", str) //--------------------------

      if r.ReceivedHandshakeMessage < r.HandshakeMessage {
        r.ReceivedHandshakeMessage++
        r.HandshakeChan <- str

        if r.ReceivedHandshakeMessage == r.HandshakeMessage {
          for _, msg := range r.Sent {
            fmt.Fprint(r.Rw, msg)
            r.Rw.Flush()
          }
        }

        continue
      }

      if r.Offset > 0 {
        r.Offset --

        continue
      }

      r.Received++
      r.ReadChan <- str
    }
  }()
}

func (r *BasicRemote)StreamHandler() (network.StreamHandler, error) {
  return func(stream network.Stream) {

    fmt.Println("[Remote] Streamhandler called ") //--------------------------

    r.Reset(bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream)))
  }, nil
}

func (r *BasicRemote)Check() bool {
  return r.Standard.Check()
}

func (r *BasicRemote)Stream() *bufio.ReadWriter {
  return r.Rw
}


func (r *BasicRemote)Close() error {
  if r.Check() {
    r.Standard.Close()
  }
  return nil
}

func (r *BasicRemote)CloseChan() chan bool {
  return r.Standard.CloseChan()
}

func (r *BasicRemote)ErrorChan() chan error {
  return r.Standard.ErrorChan()
}
