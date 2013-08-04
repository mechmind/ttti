package game


type PlayerGame struct {
    gameConnector
}


func NewPlayerGame() *PlayerGame {
    return &PlayerGame{makeGameConnector()}
}

func (p *PlayerGame) Start() {
    go p.start()
}

func (p *PlayerGame) start() {
    for
}
