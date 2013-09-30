package session

import (
    "github.com/mechmind/ttti-server/game"
    "github.com/mechmind/ttti-server/message"
)


func (s *Session) player2glyph(p string) game.Glyph {
    switch p {
    case s.p1.id:
        return game.P1_GLYPH
    case s.p2.id:
        return game.P2_GLYPH
    default:
        return game.EMPTY_GLYPH
    }
}

func (s *Session) glyph2player(g game.Glyph) string {
    switch g {
    case game.P1_GLYPH:
        return s.p1.id
    case game.P2_GLYPH:
        return s.p2.id
    default:
        return ""
    }
}


func (s *Session) makeState() (message.MsgGameState, game.Glyph) {
    field, turn, winner := s.game.GetStatus()
    msg := message.MsgGameState{"game-state", "", STATES[s.state], "", nil}
    buf := make([]byte, len(field))
    for i := range field {
        buf[i] = byte(field[i])
    }
    msg.Field = string(buf)
    msg.Turn = string([]byte{byte(turn)})
    msg.Players = make([]message.Player, 2)
    if s.p1 == nil {
        msg.Players = msg.Players[:0]
    } else {
        msg.Players[0].Name = "player1" // FIXME: player names
        msg.Players[0].Glyph = string(byte(game.P1_GLYPH))

        if s.p2 == nil {
            msg.Players = msg.Players[:1]
        } else {
            msg.Players[1].Name = "player2" // FIXME: player names
            msg.Players[1].Glyph = string(byte(game.P2_GLYPH))
        }
    }
    return msg, winner
}
