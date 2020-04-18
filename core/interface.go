package core

import (
  "log"
  "fmt"
  "os/exec"
  "bufio"
  "strings"
  "strconv"
  "errors"
  "io"
)

var (
  HeaderNotUnderstood = errors.New("Header not understood")
  CommandNotUnderstood = errors.New("Command not understood")
  //NotMatserComm = errors.New("Not the MasterComm")
  NotEnoughFields = errors.New("Not enough field")
  EmptyString = errors.New("Received an empty string")

  nilMessageHandler = func(int, string) {}
  nilRequestHandler = func(int) {}
  nilResetHandler = func(int) {}

  InterfaceLogHeader = "Log"
  InterfaceSendHeader = "Send"
  InterfaceResetHeader = "Reset"
  InterfaceRequestHeader = "Req"

  logFormat = "\033[33m%s\033[0m\n"
  masterLogFormat = "\033[32m%s\033[0m\n"
)

func NewInterface(file string, n int, i int, args ...string) (Interface, error) {
  cmdArgs := append([]string{file + "/run.py", fmt.Sprint(n), fmt.Sprint(i)}, args...)
  inter := StdInterface {
    Idx: i,
    Cmd: exec.Command("python3", cmdArgs...),
    MessageHandler: &nilMessageHandler,
    RequestHandler: &nilRequestHandler,
    ResetHandler: &nilResetHandler,
    Standard: NewStandardInterface(),
  }

  return &inter, nil
}

type StdInterface struct {
  Stdin io.Writer
  MessageHandler *func(int, string)
  RequestHandler *func(int)
  ResetHandler *func(int)
  Idx int
  Cmd *exec.Cmd
  Standard standardFunctionsCloser
}

func (s *StdInterface)Start() {
  var err error

  s.Stdin, err = s.Cmd.StdinPipe()
	if err != nil {
    s.Raise(err)
    return
	}

  stdout, err := s.Cmd.StdoutPipe()
	if err != nil {
    s.Raise(err)
    return
	}

  stderr, err := s.Cmd.StderrPipe()
	if err != nil {
    s.Raise(err)
    return
	}

  err = s.Cmd.Start()
  if err != nil {
    s.Raise(err)
    return
  }

  go func() {
    err := s.Cmd.Wait()
    if err != nil {
      s.Raise(err)
    }

    s.Close()
  }()

  errScanner := bufio.NewScanner(stderr)
  go func() {
    for errScanner.Scan() {
      s.Raise(errors.New(errScanner.Text()))
      continue
    }

    if err := errScanner.Err(); err != nil {
      s.Raise(err)
    }
  }()

  scanner := bufio.NewScanner(stdout)
  go func(){
    for s.Check() && scanner.Scan() {
      splitted := strings.Split(scanner.Text(), ",")

      switch splitted[0] {
      default:
        s.Raise(HeaderNotUnderstood)

      case InterfaceRequestHeader:
        if len(splitted) != 2 {
          s.Raise(NotEnoughFields)
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.RequestHandler)(idx)

      case InterfaceResetHeader:
        if len(splitted) != 2 {
          s.Raise(NotEnoughFields)
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.ResetHandler)(idx)

      case InterfaceLogHeader:
        if len(splitted) < 2 {
          s.Raise(NotEnoughFields)
          continue
        }

        if s.Idx == 0 {
          log.Printf(masterLogFormat, strings.Join(splitted[1:], ","))
        } else {
          log.Printf(logFormat, strings.Join(splitted[1:], ","))
        }

      case InterfaceSendHeader:
        if len(splitted) < 3 {
          s.Raise(NotEnoughFields)
          continue
        }

        idx, err := strconv.Atoi(splitted[1])
        if err != nil {
          s.Raise(err)
          continue
        }

        (*s.MessageHandler)(idx, strings.Join(splitted[2:], ","))
      }
    }

    if err := scanner.Err(); err != nil {
      s.Raise(err)
    }
  }()
}

func (s *StdInterface)Close() error {
  return s.Standard.Close()
}

func (s *StdInterface)SetErrorHandler(handler func(error)) {
  s.Standard.SetErrorHandler(handler)
}

func (s *StdInterface)SetCloseHandler(handler func()) {
  s.Standard.SetCloseHandler(handler)
}

func (s *StdInterface)Raise(err error) {
  s.Standard.Raise(err)
}

func (s *StdInterface)Check() bool {
  return s.Standard.Check()
}

func (s *StdInterface)SetMessageHandler(handler func(int, string)) {
  s.MessageHandler = &handler
}

func (s *StdInterface)SetRequestHandler(handler func(int)) {
  s.RequestHandler = &handler
}

func (s *StdInterface)SetResetHandler(handler func(int)) {
  s.ResetHandler = &handler
}

func (s *StdInterface)Push(msg string) error {
  if !s.Check() {
    return errors.New("Interface closed")
  }
  fmt.Fprintln(s.Stdin, msg)
  return nil
}
