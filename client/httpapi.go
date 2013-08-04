package main

import (
    "net/http"
    "encoding/json"
    "log"
    "io/ioutil"
    "errors"
    "net/url"
)

type HttpClient struct {
    server string
}

func NewHttpClient(server string) *HttpClient {
    return &HttpClient{server}
}


type HttpAdminClient struct {
    server string
}

func NewHttpAdminClient(server string) *HttpAdminClient {
    return &HttpAdminClient{server}
}

func (hc *HttpAdminClient) CreateSession() (string, error) {
    url := "http://" + hc.server + "/create_session"
    resp, err := http.Get(url)
    if err != nil {
        log.Println("http-admin-client: unable to create session", err)
        return "", err
    }
    defer resp.Body.Close()
    answer, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("http-admin-client: unable to read response", err)
        return "", err
    }
    var data = make(map[string]string)
    err = json.Unmarshal(answer, &data)
    if err != nil {
        log.Println("http-admin-client: unable to parse json", err)
        return "", err
    }
    resp_type, ok := data["type"]
    if !ok {
        log.Println("http-admin-client: invalid json")
        return "", errors.New("invalid json")
    }
    if resp_type == "error" {
        log.Println("http-admin-client: server error", data["message"])
        return "", errors.New(data["message"])
    }
    if resp_type != "response" {
        log.Println("http-admin-client: invalid json type")
        return "", errors.New("invalid json type")
    }
    session_id, ok := data["session_id"]
    if !ok {
        log.Println("http-admin-client: invalid json - no session_id")
        return "", errors.New("invalid json - no session_id")
    }
    return session_id, nil
}

func (hc *HttpAdminClient) AttachPlayer(session string) (string, error) {
    s_url := "http://" + hc.server + "/attach_player"
    resp, err := http.PostForm(s_url, url.Values{"session_id": {session}})
    if err != nil {
        log.Println("http-admin-client: unable to attach player", err)
        return "", err
    }
    defer resp.Body.Close()
    answer, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("http-admin-client: unable to read response", err)
        return "", err
    }
    var data = make(map[string]string)
    err = json.Unmarshal(answer, &data)
    if err != nil {
        log.Println("http-admin-client: unable to parse json", err)
        return "", err
    }
    resp_type, ok := data["type"]
    if !ok {
        log.Println("http-admin-client: invalid json")
        return "", errors.New("invalid json")
    }
    if resp_type == "error" {
        log.Println("http-admin-client: server error", data["message"])
        return "", errors.New(data["message"])
    }
    if resp_type != "response" {
        log.Println("http-admin-client: invalid json type")
        return "", errors.New("invalid json type")
    }
    session_id, ok := data["session_id"]
    if !ok {
        log.Println("http-admin-client: invalid json - no session_id")
        return "", errors.New("invalid json - no session_id")
    }
    if session_id != session {
        log.Println("http-admin-client: invalid response - session id mismatch")
        return "", errors.New("invalid response - session id mismatch")
    }
    player_id, ok := data["player_id"]
    if !ok {
        log.Println("http-admin-client: invalid json - no player_id")
        return "", errors.New("invalid json - no player_id")
    }
    return player_id, nil
}

