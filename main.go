package main

import (
    "github.com/mechmind/ttti-server/session"
)

func main() {
    registy := session.NewRegistry()
    shutdown := session.ServeHttp(":8080", registy)
    go session.ServeGame(":9090", registy)
    <-shutdown
}
