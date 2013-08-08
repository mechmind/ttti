package main

import (
    "log"

    "github.com/mechmind/ttti-server/message"
)


type Game struct {}

func (g *Game) HandleMessage(message.Message) message.Message {
    return nil
}

func runGame(c *Client) error {
    c.Start()
    game := &Game{}
    for msg := range c.connection.Read {
        log.Println("game: recieved message: ", msg)

        if msg.GetType() == "ping" {
            pong := message.MsgPong{"pong"}
            c.connection.Write <- pong
        } else if msg.GetType() == "pong" {
        } else {
            answer := game.HandleMessage(msg)
            if answer != nil {
                c.connection.Write <- answer
            }
        }
    }
    return nil
}
