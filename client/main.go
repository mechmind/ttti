package main

import (
    "log"
    "flag"

)

var session = flag.String("session", "", "session to attach to")
var server = flag.String("server", "localhost:8080", "admin interface endoint")
var host = flag.String("host", "localhost:9090", "game host endpoint")

func main() {
    flag.Parse()
    hc := NewHttpAdminClient(*server)
    var err error
    if *session == "" {
        *session, err = hc.CreateSession()
        if err != nil {
            log.Println("main: failed to create session", err)
            return
        }
        log.Println("main: created session", *session)
    }
    player, glyph, err := hc.AttachPlayer(*session)
    if err != nil {
        log.Println("main: failed to attach to session", err)
        return
    }
    log.Printf("main: attached to session '%s' as player '%s', glyph %s", *session, player, glyph)

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

