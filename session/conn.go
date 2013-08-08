package session

import (
    "net"
    "fmt"
    "log"

    "github.com/mechmind/ttti-server/message"
    "github.com/mechmind/ttti-server/connection"
)

func sendError(p *connection.PlayerConnection, code int, msg string) error {
    errmsg := message.MsgError{"error", code, msg}
    err := p.WriteMessage(errmsg)
    return err
}

func doAck(r *Registry, socket *net.TCPConn) error {
    pc := connection.NewPlayerConnection(socket)
    msg, err := pc.ReadMessage()
    if err != nil {
        log.Println("acker: cannot read greeting: ", err)
        return socket.Close()
    }
    if msg.GetType() != "attach" {
        log.Println("acket: client wont attach, weird")
        sendError(pc, 100, "first message must be 'attach'")
        return socket.Close()
    }
    greet := msg.(*message.MsgAttach)
    session := r.GetSession(greet.Sid)
    if session == nil {
        log.Println("acker: invalid session", greet.Sid)
        sendError(pc, 101, "session does not exists")
        return socket.Close()
    }
    player := session.GetPlayer(greet.Pid)
    if player == nil {
        log.Println("acker: invalid player for session", greet.Sid, greet.Pid)
        sendError(pc, 102, "no such player for this session")
        return socket.Close()
    }
    hello := &message.MsgHello{"hello"}
    err = pc.WriteMessage(hello)
    if err != nil {
        log.Println("acker: client lost after greeting", greet.Sid, greet.Pid)
        return socket.Close()
    }
    log.Printf("acker: handshake succeed for session %s, client %s", greet.Sid, greet.Pid)
    session.handleConnection(player, pc)
    return nil
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
