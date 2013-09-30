package game

import (
    "github.com/mechmind/ttti-server/message"
)

const (
    STATE_WAITING = iota
    STATE_RUNNING
)

type Game struct {
    state int
    myGlyph byte
    field Field
}

func NewGame(glyph byte) *Game {
    return &Game{0, glyph, NewField()}
}


func (g *Game) HandleMessage(m message.Message) message.Message {
    return m
}
