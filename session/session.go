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
    SESSION_WAITING = iota
    SESSION_RUNNING
    SESSION_FINISHED
    SESSION_ABORTED
)
var STATES = []string{"waiting", "running", "finished", "aborted"}

const (
    GAMEOVER_WIN = iota
    GAMEOVER_P1_LOST
    GAMEOVER_P2_LOST
    GAMEOVER_P1_SURRENDER
    GAMEOVER_P2_SURRENDER
    GAMEOVER_DRAW
)
var GAMEOVERS = []string{"win", "player1-lost", "player2-lost", "player1-surrendered",
    "player2-surrendered", "draw"}


type Session struct {
    id string
    state int
    p1, p2 *Player

    newPlayers chan *newPlayer
    regPlayer chan regPlayerReq
    ticker *time.Ticker
    game *game.Game
}

func NewSession(id string) *Session {
    return &Session{id,SESSION_WAITING,
                        &Player{"", connection.NewEmptyPlayerConnection(), time.Time{}},
                        &Player{"", connection.NewEmptyPlayerConnection(), time.Time{}},
                    make(chan *newPlayer),
                    make(chan regPlayerReq),
                    time.NewTicker(TICK_PERIOD),
                    game.NewGame()}
}

type Player struct {
    id string
    conn *connection.PlayerConnection
    lastActivity time.Time
}

func (p *Player) Send(m message.Message) {
    if p.conn.Alive {
        select {
        case p.conn.Write <- m:
        default:
        }
    }
}

type newPlayer struct {
    player *Player
    conn *connection.PlayerConnection
}

type regPlayerReq struct {
    playerId string
    reply chan regPlayerResp
}

type regPlayerResp struct {
    playerGlyph string
    err error
}

func (s *Session) AttachPlayer(playerId string) (string, error) {
    r := regPlayerReq{playerId, make(chan regPlayerResp, 1)}
    s.regPlayer <- r
    resp := <-r.reply
    return resp.playerGlyph, resp.err
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

func (s *Session) getOpponent(p *Player) *Player {
    switch {
    case s.p1 == p:
        return s.p2
    case s.p2 == p:
        return s.p1
    default:
        return nil
    }
}

func (s *Session) HandleMessage(p *Player, m message.Message) error {
    switch m.GetType() {
    case "pong":
        // ok
        log.Println("session: got pong")
    case "ping":
        // send pong
        p.Send(message.MsgPong{"pong"})
    case "make-turn":
        // check and reply 'turn'
        turn := m.(message.MsgMakeTurn)
        glyph, pos, err := checkTurn(turn)
        if err != nil {
            return err
        }
        winner, err := s.game.MakeTurn(glyph, pos)
        if err != nil {
            return err
        }
        // send turn 
        sameTurn := message.MsgTurn{"turn", turn.Coord, turn.Glyph, false}
        p.Send(sameTurn)
        oppTurn := sameTurn
        oppTurn.YouNext = true
        s.getOpponent(p).Send(oppTurn)

        if winner != game.EMPTY_GLYPH {
            sameTurn := message.MsgTurn{"turn", turn.Coord, string(byte(game.EMPTY_GLYPH)), false}
            p.Send(sameTurn)
            s.getOpponent(p).Send(sameTurn)
            // we have a winner
            s.handleWinner(winner, GAMEOVER_WIN)
        } else {
            sameTurn := message.MsgTurn{"turn", turn.Coord, turn.Glyph, false}
            p.Send(sameTurn)
            oppTurn := sameTurn
            oppTurn.YouNext = true
            s.getOpponent(p).Send(oppTurn)
        }

    default:
        log.Println("session: Got message with unknown type: ", m.GetType())
    }
    return nil
}

func (s *Session) handleWinner(winner game.Glyph, reason int) {
    s.state = SESSION_FINISHED
    winMsg := message.MsgGameOver{"game-over", string(byte(winner)), GAMEOVERS[reason]}
    s.p1.Send(winMsg)
    s.p2.Send(winMsg)
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

func (s *Session) handleLost(p *Player) {
    log.Println("session: client lost")
}

func (s *Session) handleConnect(player *Player, conn *connection.PlayerConnection) {
    err := player.conn.Close()
    if err != nil {
        log.Println("session: previous conn closed with error: ", err)
    }
    player.conn = conn
    conn.Start()
    // send state
    gameState, _ := s.makeState()
    player.Send(gameState)
    player.lastActivity = time.Now()
}

func (s *Session) startGame() {
    s.state = SESSION_RUNNING
    gameState, _ := s.makeState()
    s.p1.Send(gameState)
    s.p2.Send(gameState)
}

func (s *Session) Run() {
    for {
        select {
        case regp := <-s.regPlayer:
            if s.p1.id == "" || s.p2.id == "" {
                var glyph game.Glyph
                switch {
                case s.p1.id == "":
                    s.p1.id = regp.playerId
                    glyph = game.P1_GLYPH
                case s.p2.id == "":
                    s.p2.id = regp.playerId
                    glyph = game.P2_GLYPH
                    s.startGame()
                }
                regp.reply <- regPlayerResp{string(byte(glyph)), nil}
            } else {
                regp.reply <- regPlayerResp{"", errors.New("Session full")}
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
            err := s.HandleMessage(s.p1, m1)
            if err != nil {
                s.p1.conn.Close()
            } else {
                s.p1.lastActivity = time.Now()
            }
        case m2 := <-s.p2.conn.Read:
            err := s.HandleMessage(s.p2, m2)
            if err != nil {
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


func checkTurn(m message.MsgMakeTurn) (game.Glyph, int, error) {
    if len(m.Glyph) != 1 {
        return 0, 0, errors.New("checkTurn: glyph must be 1-character string")
    }

    glyph := game.Glyph(m.Glyph[0])
    if glyph != game.P1_GLYPH || glyph != game.P2_GLYPH {
        return 0, 0, errors.New("checkTurn: given glyph does not belong to player")
    }

    pos := m.Coord
    if pos < 0 || pos >= game.TOTAL_CELLS {
        return 0, 0, errors.New("checkTurn: pos out of range")
    }
    return glyph, int(pos), nil
}
