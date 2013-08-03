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
    player, err := hc.AttachPlayer(*session)
    if err != nil {
        log.Println("main: failed to attach to session", err)
        return
    }
    log.Printf("main: attached to session '%s' as player '%s'", *session, player)

    c := NewClient(*host, *session, player)
    err = c.Connect()
    if err != nil {
        log.Println("main: cannot establish connection", err)
    }

    log.Println("main: finished")
}
