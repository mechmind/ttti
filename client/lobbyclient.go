package main

import (
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "net/url"
)


type LobbyClient struct {
    server string
}

func NewLobbyClient(server string) *LobbyClient {
    return &LobbyClient{server}
}

type LoginMessage struct {
    Type string `json:"type"`
    Server string `json:"server"`
    Session string `json:"session_id"`
    Player string `json:"player_id"`
    Glyph string `json:"player_glyph"`
}

func (l *LobbyClient) NewGame(name string) (server, session, player, glyph string, err error) {
    serverUrl := "http://" + l.server + "/games/"
    values := url.Values{"name": {name}}
    resp, err := http.PostForm(serverUrl, values)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    answer, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return
    }

    login := &LoginMessage{}
    err = json.Unmarshal(answer, login)
    if err != nil {
        err = errors.New(string(answer) + "\n" + err.Error())
        return
    }

    // check response
    if login.Type != "join-game" {
        err = errors.New("invalid type: " + login.Type)
    }

    // all fields must be filled
    if login.Server == "" ||
        login.Session == "" ||
        login.Player == "" ||
        login.Glyph == "" {

        err = errors.New("invalid login msg")
        return
    }

    return login.Server, login.Session, login.Player, login.Glyph, nil
}
