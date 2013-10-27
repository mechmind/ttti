package game

import (
    "errors"
    "fmt"
    "io"
    "os"
    //"log"
)


const (
    STATE_WAITING = iota
    STATE_RUNNING
)

const (
    P1_GLYPH = Glyph('X')
    P2_GLYPH = Glyph('0')
    EMPTY_GLYPH = Glyph(' ')
    DRAW_GLYPH = Glyph('=')
)

const SIZE = 3
const BLOCK_SIZE = SIZE * SIZE
const TOTAL_CELLS = BLOCK_SIZE * BLOCK_SIZE

type Game struct {
    turn Glyph
    turnSquare int
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
        if g1 == EMPTY_GLYPH {
            return false, g1
        }
        return true, g1
    } else {
        return false, EMPTY_GLYPH
    }
}

func (b BigSquare) CheckWinner() Glyph {
    // check columns
    for i := 0; i < SIZE; i++ {
        if won, winner := checkWinner(b[i], b[i + SIZE], b[i + SIZE * 2]); won {
            //log.Println("won on column", i)
            return winner
        }
    }
    // check rows
    for i := 0; i < SIZE; i++ {
        if won, winner := checkWinner(b[i * SIZE], b[i * SIZE + 1], b[i * SIZE + 2]); won {
            //log.Println("won on row", i)
            return winner
        }
    }
    // check diagonals
    if won, winner := checkWinner(b[0], b[4], b[8]); won {
        //log.Println("won on 1 diag")
        return winner
    }
    if won, winner := checkWinner(b[2], b[4], b[6]); won {
        //log.Println("won on 2 diag")
        return winner
    }

    return EMPTY_GLYPH
}

func (b BigSquare) HasEmpty() bool {
    for _, cell := range b {
        if cell == EMPTY_GLYPH {
            return true
        }
    }
    return false
}

func NewField() *Field {
    field := Field{make([]BigSquare, BLOCK_SIZE), make([]Glyph, BLOCK_SIZE)}
    for i := 0; i < BLOCK_SIZE; i++ {
        field.squares[i] = make([]Glyph, BLOCK_SIZE)
        for j := 0; j < BLOCK_SIZE; j++ {
            field.squares[i][j] = EMPTY_GLYPH
        }
    }

    for i := 0; i < BLOCK_SIZE; i++ {
        field.won[i] = EMPTY_GLYPH
    }

    return &field
}

func parseField(state string) (*Field, error) {
    field := NewField()
    for i := 0; i < BLOCK_SIZE; i++ {
        for j := 0; j < BLOCK_SIZE; j++ {
            cell := Glyph(state[i * BLOCK_SIZE + j])
            if cell != P1_GLYPH && cell != P2_GLYPH && cell != EMPTY_GLYPH {
                return nil, fmt.Errorf("game: invalid field symbol at %d", i * BLOCK_SIZE + j)
            }
            field.squares[i][j] = cell
        }
    }
    return field, nil
}

func NewGame() *Game {
    return &Game{P1_GLYPH, 0, NewField()} // FIXME: starting square
}

func LoadGame(turn Glyph, turnSquare int, state string) (*Game, error) {
    if turn != P1_GLYPH && turn != P2_GLYPH {
        return nil, errors.New("game: invalid turn in LoadGame")
    }
    if len(state) != TOTAL_CELLS {
        return nil, errors.New("game: invalid state length")
    }

    if turnSquare < 0 || turnSquare >= BLOCK_SIZE {
        return nil, errors.New("game: invalid turn square")
    }

    field, err := parseField(state)
    if err != nil {
        return nil, err
    }
    return &Game{turn, turnSquare, field}, nil
}

func (g *Game) MakeTurn(gl Glyph, pos int) (next Glyph, nextSquare int, winner Glyph, err error) {
    winner = g.field.won.CheckWinner()
    if winner != EMPTY_GLYPH {
        err = errors.New("game: game over, no turns, winner is " + string(byte(winner)))
        //err = fmt.Errorf("game: game over, no turns, winner is '%b'", winner)
        return
    }
    if g.turn != gl {
        err = errors.New("game: invalid turn")
        return
    }

    block, cell := g.splitPos(pos)
    if block != g.turnSquare {
        err = fmt.Errorf("game: invalud turn square, expected '%d', got '%d'", g.turnSquare, block)
        //err = errors.New("game: invalid turn square")
        return
    }

    if g.field.squares[block][cell] != EMPTY_GLYPH {
        g.DumpTo(os.Stdout)
        //err = errors.New("game: invalid turn - cell already occupied")
        return
    }

    g.field.squares[block][cell] = gl
    g.field.won[block] = g.field.squares[block].CheckWinner()

    if g.field.squares[cell].HasEmpty() {
        nextSquare = cell
    } else {
        // if no free squares, its a gameover
        nextSquare = -1
        // find square with free cells
        for idx, sq := range g.field.squares {
            if sq.HasEmpty() {
                nextSquare = idx
                break
            }
        }
    }

    if gl == P1_GLYPH {
        next = P2_GLYPH
    } else {
        next = P1_GLYPH
    }

    winner = g.field.won.CheckWinner()
    if winner != EMPTY_GLYPH {
        next = EMPTY_GLYPH
        nextSquare = -1
    }

    if winner == EMPTY_GLYPH && nextSquare == -1 {
        // no winner and no available blocks -> draw
        winner = DRAW_GLYPH
    }

    g.turnSquare = nextSquare
    g.turn = next
    return
}

func (g *Game) splitPos(pos int) (block int, cell int) {
    return pos / BLOCK_SIZE, pos % BLOCK_SIZE
}

func (g *Game) GetStatus() (field []Glyph, turn Glyph, turnSquare int, win Glyph) {
    field = make([]Glyph, TOTAL_CELLS)
    for i := 0; i < BLOCK_SIZE; i++ {
        copy(field[BLOCK_SIZE * i:], g.field.squares[i])
    }
    return field, g.turn, g.turnSquare, g.field.won.CheckWinner()
}

func (g *Game) GetSquare(id int) BigSquare {
    square := make(BigSquare, BLOCK_SIZE)
    copy(square, g.field.squares[id])
    return square
}

func (g *Game) GetTurnId() int {
    return g.turnSquare
}

func (g *Game) DumpTo (w io.Writer) error {
    var dataBuf = []byte("|   |   |   |\n")
    var lineBuf = []byte("+---+---+---+\n")

    for line := 0; line < BLOCK_SIZE + 4; line++ {
        if line % 4 == 0 {
            _, err := w.Write(lineBuf)
            if err != nil {
                return err
            }
        } else {
            blockLevel := line / 4 * 3
            rowLevel := (line - 1) % 4 * 3
            for i := 0; i < 3; i++ {
                for j := 0; j < 3; j++ {
                    dataBuf[i * 4 + 1 + j] = byte(g.field.squares[i + blockLevel][j + rowLevel])
                }
            }
            _, err := w.Write(dataBuf)
            if err != nil {
                return err
            }
        }
    }
    return nil
}
