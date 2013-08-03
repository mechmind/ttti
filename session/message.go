package session

import (
    "encoding/json"
    "errors"
)

type Message interface {
    GetType() string
}

func parseMessage(src []byte) (Message, error) {
    var parsed interface{}
    // first, parse as map and extract `type`
    err := json.Unmarshal(src, &parsed)
    if err != nil {
        return nil, err
    }

    data, ok := parsed.(map[string]interface{})
    if !ok {
        return nil, errors.New("Invalid packet: must be an object")
    }

    msgTypeIface, ok := data["type"]
    if !ok {
        return nil, errors.New("Invalid packet: missing type")
    }

    msgType, ok := msgTypeIface.(string)
    if !ok {
        return nil, errors.New("Invalid packet: `type` must be a string")
    }

    var m Message
    switch msgType {
    case "attach":
        m = &MsgAttach{}
    }

    err = json.Unmarshal(src, m)
    if err != nil {
        return nil, err
    }

    return m, nil
}

func serializeMessage(m Message) ([]byte, error) {
    return json.Marshal(m)
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

type MsgError struct {
    Type `json:"type"`
    Code int `json:"code"`
    Message string `json:"message"`
}

