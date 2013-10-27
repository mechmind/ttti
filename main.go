package main

import (
    "flag"

    "github.com/mechmind/ttti-server/session"
)

var apiEndpoint = flag.String("api", "127.0.0.1:9091", "admin api endpoint")
var gameEndpoint = flag.String("game", "127.0.0.1:9090", "game client endpoint")


func main() {
    flag.Parse()

    registy := session.NewRegistry()
    shutdown := session.ServeHttp(*apiEndpoint, registy)
    go session.ServeGame(*gameEndpoint, registy)
    <-shutdown
}
