package main

import (
    "log"
    "flag"

)

var session = flag.String("session", "", "session to attach to")
var server = flag.String("server", "", "admin interface endoint")
var host = flag.String("host", "", "game host endpoint")

var lobby = flag.String("lobby", "localhost:8080", "lobby api endpoint")
var name = flag.String("name", "player", "player name")


func main() {
    flag.Parse()
    var err error
    var player, glyph string

    if *host != "" && *server != "" {
        // use admin interface directly
        log.Println("main: admin-mode")
        hc := NewHttpAdminClient(*server)
        if *session == "" {
            *session, err = hc.CreateSession()
            if err != nil {
                log.Println("main: failed to create session", err)
                return
            }
            log.Println("main: created session", *session)
        }
        player, glyph, err = hc.AttachPlayer(*session)
        if err != nil {
            log.Println("main: failed to attach to session", err)
            return
        }
        log.Printf("main: attached to session '%s' as player '%s', glyph %s", *session, player,
            glyph)
    } else {
        // use lobby registrator
        log.Println("main: client-mode")
        client := NewLobbyClient(*lobby)
        *host, *session, player, glyph, err = client.NewGame(*name)
        if err != nil {
            log.Println("main: cannot join game: ", err)
            return
        }
    }

    c := NewClient(*host, *session, player, glyph)
    err = c.Connect()
    if err != nil {
        log.Println("main: cannot establish connection", err)
        return
    }

    err = runGame(c)
    if err != nil {
        log.Println("main: game finished with error: ", err)
    }

    log.Println("main: finished")
}

