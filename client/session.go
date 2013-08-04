package main

import (
    "errors"
    "net"

    "github.com/mechmind/ttti-server/connection"
    "github.com/mechmind/ttti-server/message"
)


type Client struct {
    host string
    session string
    player string
    connection *connection.PlayerConnection
}

func NewClient(host string, session string, player string) *Client {
    return &Client{host, session, player, nil}
}

func (c *Client) Connect() error {
    // open socket
    addr, err := net.ResolveTCPAddr("tcp", c.host)
    if err != nil {
        return err
    }
    socket, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return err
    }
    conn, err := c.handshake(socket)
    if err != nil {
        socket.Close()
        return err
    }
    c.connection = conn
    return nil
}

func (c *Client) handshake(socket *net.TCPConn) (*connection.PlayerConnection, error) {
    attach := &message.MsgAttach{"attach", c.session, c.player}
    conn := connection.NewPlayerConnection(socket)
    err := conn.WriteMessage(attach)
    if err != nil {
        return nil, err
    }

    hello, err := conn.ReadMessage()
    if err != nil {
        return nil, err
    }
    if hello.GetType() != "hello" {
        switch hello.GetType() {
        case "error":
            errmsg := hello.(message.MsgError)
            return nil, errors.New("server kiks us with error: " + errmsg.Message)
        default:
            return nil, errors.New("server going mad! Type is: " + hello.GetType())
        }
    }
    return conn, nil
}

