package session

import (
    "net"
    "fmt"
    "log"
    "bufio"
    "encoding/json"
)

type Greeting struct {
    Type string `json:"type"`
    Sid string `json:"sid"`
    Pid string `json:"pid"`
}

func sendError(socket *net.TCPConn, code int) error {
    _, err := fmt.Fprintf(socket, `{"type": "error", "code": %d, "fatal": true, "message": ""}\n`, code)
    if err != nil {
        return err
    }
    return nil
}

func doAck(r *Registry, socket *net.TCPConn) {
    reader := bufio.NewScanner(socket)
    if !reader.Scan() {
        err := reader.Err()
        log.Println("acker: cannot read ack", err)
        socket.Close()
    } else {
        header := reader.Bytes()
        var greet = Greeting{}
        err := json.Unmarshal(header, &greet)
        if err != nil {
            log.Println("acker: invalid greeting", err)
            sendError(socket, 100)
            socket.Close()
        } else {
            if greet.Type != "attach" {
                log.Println("acker: invalid greeting - not an attach")
                sendError(socket, 100)
                socket.Close()
            }
            session := r.GetSession(greet.Sid)
            if session == nil {
                log.Println("acker: invalid session", greet.Sid)
                sendError(socket, 101)
                socket.Close()
            } else {
                player := session.GetPlayer(greet.Pid)
                if player == nil {
                    log.Println("acker: invalid player for session", greet.Sid, greet.Pid)
                    sendError(socket, 102)
                    socket.Close()
                } else {
                    _, err = socket.Write([]byte("{\"type\":\"hello\"}\n"))
                    if err != nil {
                        log.Println("acker: client lost after greeting", greet.Sid, greet.Pid)
                        socket.Close()
                    } else {
                        session.handleConnection(player, socket, reader)
                    }
                }
            }
        }
    }
}

func ServeGame(addrstr string, r *Registry) {
    addr, err := net.ResolveTCPAddr("tcp", addrstr)
    if err != nil {
        panic(err)
    }
    socket, err := net.ListenTCP("tcp", addr)
    if err != nil {
        panic(err)
    }

    for {
        conn, err := socket.AcceptTCP()
        if err != nil {
            fmt.Println("Error while accepting connections:", err)
        } else {
            go doAck(r, conn)
        }
    }
}
