package session

import (
    "errors"
    "log"
    "time"

    "github.com/mechmind/ttti-server/message"
    "github.com/mechmind/ttti-server/connection"
    "github.com/mechmind/ttti-server/game"
)

const TICK_PERIOD = 1 * time.Second
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

const (
    REPLY_TO_ALL = iota
    REPLY_TO_SENDER
    REPLY_TO_ENEMY
)

type Session struct {
    id string
    p1, p2 *Player

    newPlayers chan *newPlayer
    regPlayer chan regPlayer
    ticker *time.Ticker
    game *game.Game
}

func NewSession(id string) *Session {
    return &Session{id, &Player{"", connection.NewEmptyPlayerConnection(), time.Time{}},
                        &Player{"", connection.NewEmptyPlayerConnection(), time.Time{}},
                    make(chan *newPlayer),
                    make(chan regPlayer),
                    time.NewTicker(TICK_PERIOD),
                    game.NewGame()}
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

type regPlayer struct {
    player_id string
    reply chan error
}

func (s *Session) AttachPlayer(player_id string) error {
    r := regPlayer{player_id, make(chan error, 1)}
    s.regPlayer <- r
    return <-r.reply
}

func (s *Session) handleConnection(p *Player, pc *connection.PlayerConnection) {
    newp := newPlayer{p, pc}
    s.newPlayers <- &newp
}

func (s *Session) GetPlayer(player_id string) *Player {
    switch {
    case s.p1.id == player_id:
        return s.p1
    case s.p2.id == player_id:
        return s.p2
    default:
        return nil
    }
}

func (s *Session) HandleMessage(p *Player, m message.Message) (message.Message, int, error) {
    switch m.GetType() {
    case "pong":
        // ok
    default:
        log.Println("session: Got message with unknown type: ", m.GetType())
    }
    return nil, REPLY_TO_SENDER, nil
}

func (s *Session) checkStale(p *Player, now time.Time) {
    if p.conn.Alive && now.Sub(p.lastActivity) > AWARE_TIMEOUT {
        log.Println("session: sending ping to client")
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

func (s *Session) handleConnect(player *Player, conn *connection.PlayerConnection) {
    err := player.conn.Close()
    if err != nil {
        log.Println("session: socket closed with error: ", err)
    }
    player.conn = conn
    conn.Start()
    player.lastActivity = time.Now()
}

func (s *Session) startGame() {

}

func (s *Session) Run() {
    for {
        select {
        case regp := <-s.regPlayer:
            if s.p1.id == "" || s.p2.id == "" {
                switch {
                case s.p1.id == "":
                    s.p1.id = regp.player_id
                case s.p2.id == "":
                    s.p2.id = regp.player_id
                    s.startGame()
                }
                regp.reply <- nil
            } else {
                regp.reply <- errors.New("Session full")
            }
        case newp := <-s.newPlayers:
            if newp.player == s.p1 {
                s.handleConnect(s.p1, newp.conn)
                log.Println("session: accepted new connection for player1")
            } else if newp.player == s.p2 {
                s.handleConnect(s.p2, newp.conn)
                log.Println("session: accepted new connection for player2")
            } else {
                // wtf?
                log.Print("session: got connection for unknown player", newp.player.id)
                newp.conn.Close()
            }
        case m1 := <-s.p1.conn.Read:
            msg, to, err := s.HandleMessage(s.p1, m1)
            if err != nil {
                s.p1.conn.Close()
            } else {
                s.p1.lastActivity = time.Now()
            }
            if msg != nil {
                replyTo(msg, to, s.p1.conn.Write, s.p2.conn.Write)
            }
        case m2 := <-s.p2.conn.Read:
            msg, to, err := s.HandleMessage(s.p2, m2)
            if err != nil {
                s.p2.conn.Close()
            } else {
                s.p2.lastActivity = time.Now()
            }
            if msg != nil {
                replyTo(msg, to, s.p2.conn.Write, s.p1.conn.Write)
            }
        case now := <-s.ticker.C:
            s.checkStale(s.p1, now)
            s.checkStale(s.p2, now)
        }
    }
}


func writeIfCan(ch chan message.Message, msg message.Message) {
    if ch == nil {
        return
    }
    select {
    case ch <- msg:
    default: // channel is full -> connection stale
    }
}

func replyTo(msg message.Message, to int, me, enemy chan message.Message) {
    switch to {
    case REPLY_TO_SENDER:
        writeIfCan(me, msg)
    case REPLY_TO_ENEMY:
        writeIfCan(enemy, msg)
    case REPLY_TO_ALL:
        writeIfCan(me, msg)
        writeIfCan(enemy, msg)
    }
}
