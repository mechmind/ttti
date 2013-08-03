package message

import (
    "encoding/json"
    "errors"
)


func ParseMessage(src []byte) (Message, error) {
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
    case "ping":
        m = &MsgPing{}
    case "pong":
        m = &MsgPong{}
    }

    err = json.Unmarshal(src, m)
    if err != nil {
        return nil, err
    }

    return m, nil
}

func SerializeMessage(m Message) ([]byte, error) {
    return json.Marshal(m)
}

