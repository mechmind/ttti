package session

import (
    "errors"
    "log"
    "time"

    "github.com/mechmind/ttti-server/message"
    "github.com/mechmind/ttti-server/connection"
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
    return &Session{id, &Player{"", connection.NewEmptyPlayerConnection(), time.Time{}},
                        &Player{"", connection.NewEmptyPlayerConnection(), time.Time{}},
                    make(chan *newPlayer),
                    time.NewTicker(AWARE_TIMEOUT),
                    SESSION_PRESTART}
}

type Player struct {
    id string
    conn *connection.PlayerConnection
    lastActivity time.Time
}

type newPlayer struct {
    player *Player
    conn *connection.PlayerConnection
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

func (s *Session) handleConnection(p *Player, pc *connection.PlayerConnection) {
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
    if p.conn.Alive && p.lastActivity.Sub(now) > AWARE_TIMEOUT {
        p.conn.Write <- message.MsgPing{"ping"}
    }

    if p.conn.Alive && p.lastActivity.Sub(now) > DEADLINE_TIMEOUT {
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
        case m1 := <-s.p1.conn.Read:
            if s.ProcessMessage(s.p1, m1) != nil {
                s.p1.conn.Close()
            } else {
                s.p1.lastActivity = time.Now()
            }
        case m2 := <-s.p2.conn.Read:
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


