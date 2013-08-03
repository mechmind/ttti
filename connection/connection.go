package connection

import (
    "net"
    "bufio"
    "errors"

    "tic-tac-inception-toe/message"
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
    return &PlayerConnection{make(chan message.Message), make(chan message.Message), true, false,
        socket, bufio.NewScanner(socket), make(chan bool)}
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
    for {
        message, err := p.ReadMessage()
        if err != nil {
            p.Close()
            p.closeCh <- true
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
    for {
        select {
        case msg := <-p.Write:
            err := p.WriteMessage(msg)
            if err != nil {
                p.socket.Close()
                break
            }
        case <-p.closeCh:
            break
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

