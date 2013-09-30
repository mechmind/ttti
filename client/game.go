package main

import (
    "log"
    "os"
    "time"
    "math/rand"

    "github.com/mechmind/ttti-server/message"
    "github.com/mechmind/ttti-server/game"
)

type Gamer struct {
    game *game.Game
    glyph game.Glyph
}

func runGame(c *Client) error {
    // init rand
    rand.Seed(time.Now().UnixNano())

    c.Start()
    gamer := &Gamer{nil, game.Glyph(c.glyph[0])}
    for msg := range c.connection.Read {
        log.Println("game: recieved message: ", msg)

        if msg.GetType() == "ping" {
            pong := message.MsgPong{"pong"}
            c.connection.Write <- pong
        } else if msg.GetType() == "pong" {
        } else {
            err := handleMessage(c, gamer, msg)
            if err != nil {
                log.Println("error handling message: ", err)
            }
        }
    }
    return nil
}

func handleMessage(client *Client, gamer *Gamer, msg message.Message) error {
    var err error
    switch msg.GetType() {
    case "game-state":
        // load state
        state := msg.(*message.MsgGameState)
        field := state.Field
        turn := game.Glyph(state.Turn[0])
        gamer.game, err = game.LoadGame(turn, state.TurnSquare, field)
        if err != nil {
            return err
        }
        //debug
        gamer.game.DumpTo(os.Stdout)
        // if our turn, make it
        if turn == gamer.glyph {
            makeTurn(client, gamer)
        }
    case "turn":
        // someone made turn
        turn := msg.(*message.MsgTurn)
        _, _, _, err := gamer.game.MakeTurn(game.Glyph(turn.Glyph[0]), turn.Coord)
        if err != nil {
            log.Fatal("turn: server sent invalid turn: ", err, turn)
        }

        gamer.game.DumpTo(os.Stdout)

        // if our turn, make it
        if game.Glyph(turn.NextGlyph[0]) == gamer.glyph {
            makeTurn(client, gamer)
        }
    case "game-over":
        // game is over
        won := msg.(*message.MsgGameOver)
        gamer.game.DumpTo(os.Stdout)
        if game.Glyph(won.Winner[0]) == gamer.glyph {
            log.Fatalf("game is over! I WON!!!")
        } else {
            log.Fatalf("game is over! I loose...")
        }
    case "error":
        // got error
        log.Fatal("got error from server: ", msg)
    default:
        log.Println("Unknown message: ", msg.GetType(), msg)
    }
    return nil
}

func makeTurn(client *Client, gamer *Gamer) {
    // take random free cell from given square
    sq := gamer.game.GetSquare(gamer.game.GetTurnId())
    free := make([]int, game.BLOCK_SIZE)
    var nextIdx int
    for idx, cell := range sq {
        if cell == game.EMPTY_GLYPH {
            free[nextIdx] = idx
            nextIdx++
        }
    }
    free = free[:nextIdx]
    if len(free) == 0 {
        log.Fatalln("makeTurn: WTF? No free cells")
    }

    idx := free[rand.Intn(len(free))]
    coord := gamer.game.GetTurnId() * game.BLOCK_SIZE + idx

    msg := &message.MsgMakeTurn{"make-turn", coord, string(byte(gamer.glyph))}
    log.Printf("Making turn: %c at %d [%d,%d]", byte(gamer.glyph), coord, gamer.game.GetTurnId(),
        idx)
    client.connection.Write <- msg
}
