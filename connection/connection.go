package connection

import (
    "net"
    "bufio"
    "errors"
    "log"

    "github.com/mechmind/ttti-server/message"
)

const (
    READ_BUFFER = 5
    WRITE_BUFFER = 5
)

type PlayerConnection struct {
    Read, Write chan message.Message
    Alive bool
    lost bool
    socket *net.TCPConn
    reader *bufio.Scanner
    closeCh chan bool
}

func NewEmptyPlayerConnection() *PlayerConnection {
    return &PlayerConnection{nil, nil, false, false, nil, nil, make(chan bool)}
}

func NewPlayerConnection(socket *net.TCPConn) *PlayerConnection {
    return &PlayerConnection{make(chan message.Message, READ_BUFFER),
        make(chan message.Message, WRITE_BUFFER), true, false, socket, bufio.NewScanner(socket),
        make(chan bool, 1)}
}

func (p *PlayerConnection) Close() error {
    if p.Alive {
        p.closeCh <- true
        p.Alive = false
        if p.socket != nil {
            return p.socket.Close()
        }
    }
    return nil
}

func (p *PlayerConnection) Start() {
    go p.startReadLoop()
    go p.startWriteLoop()
}

func (p *PlayerConnection) startReadLoop() {
    log.Println("connection: Starting read loop...")
    defer log.Println("connection: Read loop finished")
    for {
        message, err := p.ReadMessage()
        if err != nil {
            log.Println("connection: Read loop got error, exiting")
            p.Close()
            return
        }
        p.Read <- message
    }
}

func (p *PlayerConnection) ReadMessage() (message.Message, error) {
    ok := p.reader.Scan()
    if !ok {
        // socket closed or err
        return nil, errors.New("Error reading from socket")
    }
    line := p.reader.Bytes()
    message, err := message.ParseMessage(line)
    if err != nil {
        // TODO: proper error handling
        return nil, err
    }
    return message, nil
}

func (p *PlayerConnection) startWriteLoop() {
    log.Println("connection: Starting write loop...")
    defer log.Println("connection: Write loop finished")
    for {
        select {
        case msg := <-p.Write:
            err := p.WriteMessage(msg)
            if err != nil {
                log.Println("connection: Write loop got error, exiting")
                p.socket.Close()
                return
            }
        case <-p.closeCh:
            log.Println("connection: Write loop got exit message, exiting")
            return
        }
    }
}

func (p *PlayerConnection) WriteMessage(msg message.Message) error {
    buf, err := message.SerializeMessage(msg)
    if err != nil {
        return err
    }
    buf = append(buf, '\n')
    _, err = p.socket.Write(buf)
    if err != nil {
        return err
    }
    return nil
}

