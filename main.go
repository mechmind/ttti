package main

import (
    "balda/session"
)

func main() {
    registy := session.NewRegistry()
    shutdown := session.ServeHttp(":8080", registy)
    <-shutdown
}
