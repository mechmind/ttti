package game

import (
    "errors"
)


const (
    STATE_WAITING = iota
    STATE_RUNNING
)

const (
    P1_GLYPH = Glyph('X')
    P2_GLYPH = Glyph('0')
    EMPTY_GLYPH = Glyph(' ')
)

const SIZE = 3
const BLOCK_SIZE = SIZE * SIZE
const TOTAL_CELLS = BLOCK_SIZE * BLOCK_SIZE

type Game struct {
    turn Glyph
    field *Field
}

type Field struct {
    squares []BigSquare
    won BigSquare
}


type BigSquare []Glyph
type Glyph byte

func checkWinner(g1, g2, g3 Glyph) (bool, Glyph) {
    if g1 == g2 && g2 == g3 {
        return true, g1
    } else {
        return false, EMPTY_GLYPH
    }
}

func (b BigSquare) checkWinner() Glyph {
    // check rows
    for i := 0; i < SIZE; i++ {
        if won, winner := checkWinner(b[i], b[i + SIZE], b[i + SIZE * 2]); won {
            return winner
        }
    }
    // check columns
    for i := 0; i < SIZE; i++ {
        if won, winner := checkWinner(b[i * SIZE], b[i * SIZE + 1], b[i * SIZE + 2]); won {
            return winner
        }
    }
    // check diagonals
    if won, winner := checkWinner(b[0], b[5], b[9]); won {
        return winner
    }
    if won, winner := checkWinner(b[3], b[5], b[6]); won {
        return winner
    }

    return EMPTY_GLYPH
}

func NewField() *Field {
    field := Field{make([]BigSquare, BLOCK_SIZE), make([]Glyph, BLOCK_SIZE)}
    for i := 0; i < BLOCK_SIZE; i++ {
        field.squares[i] = make([]Glyph, BLOCK_SIZE)
        for j := 0; j < BLOCK_SIZE; j++ {
            field.squares[i][j] = EMPTY_GLYPH
        }
    }
    return &field
}

func NewGame() *Game {
    return &Game{P1_GLYPH, NewField()}
}

func (g *Game) MakeTurn(gl Glyph, pos int) (winner Glyph, err error) {
    winner = g.field.won.checkWinner()
    if winner != EMPTY_GLYPH {
        return winner, errors.New("game: game over, no turns")
    }
    if g.turn != gl {
        return EMPTY_GLYPH, errors.New("game: invalid turn")
    }

    block, cell := g.splitPos(pos)
    g.field.squares[block][cell] = gl
    g.field.won[block] = g.field.squares[block].checkWinner()
    return g.field.won.checkWinner(), nil
}

func (g *Game) splitPos(pos int) (block int, cell int) {
    return pos % BLOCK_SIZE, pos / BLOCK_SIZE
}

func (g *Game) GetStatus() (field []Glyph, turn Glyph, win Glyph) {
    field = make([]Glyph, TOTAL_CELLS)
    for i := 0; i < BLOCK_SIZE; i++ {
        copy(field[BLOCK_SIZE * i:], g.field.squares[i])
    }
    return field, g.turn, g.field.won.checkWinner()
}
