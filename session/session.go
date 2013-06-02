package session

import (
    "net"
    "errors"
)


type Session struct {
    id string
    p1, p2 *Player
}

type Player struct {
    id string
}

func (s *Session) AttachPlayer(player_id string) error {
    var player = &Player{player_id}
    switch {
    case s.p1 == nil:
        s.p1 = player
    case s.p2 == nil:
        s.p2 = player
    default:
        return errors.New("Session full")
    }
    return nil
}

type ConnectionReader struct {
    conn net.TCPConn
    shutdownChannel chan chan bool
    //sessionChannel chan *Message
}

type ConnectionWriter struct {}

