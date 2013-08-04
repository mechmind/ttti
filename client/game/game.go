package game

import (
    "github.com/mechmind/ttti-server/message"
)

type gameConnector struct {
    Input, Output chan message.Message
}

func makeGameConnector() gameConnector {
    return gameConnector{make(chan message.Message), make(chan message.Message)}
}


type Game interface {
    Start()
}
