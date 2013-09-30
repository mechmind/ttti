package main

import (
    "log"

    "github.com/mechmind/ttti-server/message"
    "github.com/mechmind/ttti-server/client/game"
)


func runGame(c *Client) error {
    c.Start()
    g := game.NewGame(c.glyph[0])
    for msg := range c.connection.Read {
        log.Println("game: recieved message: ", msg)

        if msg.GetType() == "ping" {
            pong := message.MsgPong{"pong"}
            c.connection.Write <- pong
        } else if msg.GetType() == "pong" {
        } else {
            handleMessage(g, msg)
        }
    }
    return nil
}

func handleMessage(game *game.Game, msg message.Message) error {
    switch msg.GetType() {
    case "game-state":
        // load state
    case "turn":
        // opponent made turn
    case "game-over":
        // game is over
    case "error":
        // got error
    default:
        log.Println("Unknown message: ", msg)
    }
    return nil
}
