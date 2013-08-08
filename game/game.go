package game


const (
    STATE_WAITING = iota
    STATE_RUNNING
)

const (
    P1_GLYPH = Turn('X')
    P2_GLYPH = Turn('0')
)

type Game struct {
    state int
    field Field
}

type Field [9]BigSquare
type BigSquare [9]byte
type Turn byte

func NewGame() *Game {
    return &Game{}
}

func (g *Game) Start() Turn {
    g.state = STATE_RUNNING
    return P1_GLYPH
}
