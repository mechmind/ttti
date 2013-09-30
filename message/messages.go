package message


type Message interface {
    GetType() string
}


type Type string

func (t Type) GetType() string {
    return string(t)
}


type MsgAttach struct {
    Type `json:"type"`
    Sid string `json:"sid"`
    Pid string `json:"pid"`
}

type MsgHello struct {
    Type `json:"type"`
}

type MsgPing struct {
    Type `json:"type"`
}

type MsgPong struct {
    Type `json:"type"`
}

type MsgMakeTurn struct {
    Type `json:"type"`
    Coord int32 `json:"coord"`
    Glyph string `json:"glyph"`
}

type MsgTurn struct {
    Type `json:"type"`
    Coord int32 `json:"coord"`
    Glyph string `json:"glyph"`
    YouNext bool `json:"you_next"`
}

type MsgError struct {
    Type `json:"type"`
    Code int `json:"code"`
    Message string `json:"message"`
}

type MsgGameState struct {
    Type `json:"type"`
    Field string `json:"field"`
    State string `json:"state"`
    Turn string `json:"turn"`
    Players []Player `json:"players"`
}

type MsgGameOver struct {
    Type `json:"type"`
    Winner string `json:"winner"`
    Reason string `json:"reason"`
}


type Player struct {
    Name string `json:"name"`
    Glyph string `json:"glyph"`
}
