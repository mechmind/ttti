package session

import (
    "net"
    "errors"
    "bufio"
    "log"
    "time"

    "tic-tac-inception-toe/message"
)

const AWARE_TIMEOUT = 5 * time.Second
const DEADLINE_TIMEOUT = 30 * time.Second
const LOST_TIMEOUT = 1 * time.Minute

const (
    SESSION_PRESTART = iota
    SESSION_WAITING
    SESSION_RUNNING
    SESSION_FINISHED
    SESSION_ABORTED
)

type Session struct {
    id string
    p1, p2 *Player

    newPlayers chan *newPlayer
    ticker *time.Ticker
    state int
}

func NewSession(id string) *Session {
    return &Session{id, &Player{"", newEmptyPlayerConnection(), time.Time{}},
                        &Player{"", newEmptyPlayerConnection(), time.Time{}},
                    make(chan *newPlayer),
                    time.NewTicker(AWARE_TIMEOUT),
                    SESSION_PRESTART}
}

type Player struct {
    id string
    conn *playerConnection
    lastActivity time.Time
}

type playerConnection struct {
    read, write chan message.Message
    alive bool
    lost bool
    socket *net.TCPConn
    reader *bufio.Scanner
    closeCh chan bool
}

func newEmptyPlayerConnection() *playerConnection {
    return &playerConnection{nil, nil, false, false, nil, nil, make(chan bool)}
}

func newPlayerConnection(socket *net.TCPConn, reader *bufio.Scanner) *playerConnection {
    return &playerConnection{make(chan message.Message), make(chan message.Message), true, false,
        socket, reader, make(chan bool)}
}

func (p *playerConnection) Close() error {
    if p.alive {
        p.closeCh <- true
        p.alive = false
        if p.socket != nil {
            return p.socket.Close()
        }
    }
    return nil
}

func (p *playerConnection) Start() {
    go p.startReadLoop()
    go p.startWriteLoop()
}

func (p *playerConnection) startReadLoop() {
    for {
        message, err := p.readMessage()
        if err != nil {
            p.Close()
            p.closeCh <- true
            return
        }
        p.read <- message
    }
}

func (p *playerConnection) readMessage() (message.Message, error) {
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

func (p *playerConnection) startWriteLoop() {
    for {
        select {
        case msg := <-p.write:
            err := p.writeMessage(msg)
            if err != nil {
                p.socket.Close()
                break
            }
        case <-p.closeCh:
            break
        }
    }
}

func (p *playerConnection) writeMessage(msg message.Message) error {
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

type newPlayer struct {
    player *Player
    conn *playerConnection
}

func (s *Session) AttachPlayer(player_id string) error {
    switch {
    case s.p1.id == "":
        s.p1.id = player_id
    case s.p2.id == "":
        s.p2.id = player_id
    default:
        return errors.New("Session full")
    }
    return nil
}

func (s *Session) handleConnection(p *Player, pc *playerConnection) {
    newp := newPlayer{p, pc}
    s.newPlayers <- &newp
}

func (s *Session) GetPlayer(player_id string) *Player {
    switch {
    case s.p1.id == player_id:
        return s.p1
    case s.p1.id == player_id:
        return s.p2
    default:
        return nil
    }
}

func (s *Session) ProcessMessage(p *Player, m message.Message) error {
    return nil
}

func (s *Session) checkStale(p *Player, now time.Time) {
    if p.conn.alive && p.lastActivity.Sub(now) > AWARE_TIMEOUT {
        p.conn.write <- message.MsgPing{"ping"}
    }

    if p.conn.alive && p.lastActivity.Sub(now) > DEADLINE_TIMEOUT {
        p.conn.Close()
    }

    if p.lastActivity.Sub(now) > LOST_TIMEOUT {
        s.handleLost(p)
    }
}

func (s *Session) handleLost(p *Player) {}

func (s *Session) Run() {
    for {
        select {
        case newp := <-s.newPlayers:
            if newp.player == s.p1 {
                err := s.p1.conn.Close()
                if err != nil {
                    log.Println("session: socket closed with error: ", err)
                }
                s.p1.conn = newp.conn
            } else if newp.player == s.p2 {
                err := s.p2.conn.Close()
                if err != nil {
                    log.Println("session: socket closed with error: ", err)
                }
                s.p2.conn = newp.conn
            } else {
                // wtf?
                log.Print("session: got connection for unknown player", newp.player.id)
                newp.conn.Close()
            }
        case m1 := <-s.p1.conn.read:
            if s.ProcessMessage(s.p1, m1) != nil {
                s.p1.conn.Close()
            } else {
                s.p1.lastActivity = time.Now()
            }
        case m2 := <-s.p2.conn.read:
            if s.ProcessMessage(s.p2, m2) != nil {
                s.p2.conn.Close()
            } else {
                s.p2.lastActivity = time.Now()
            }
        case now := <-s.ticker.C:
            s.checkStale(s.p1, now)
            s.checkStale(s.p2, now)
        }
    }
}


