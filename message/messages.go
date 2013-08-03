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

type MsgError struct {
    Type `json:"type"`
    Code int `json:"code"`
    Message string `json:"message"`
}

