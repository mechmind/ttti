package session

import (
    "sync"
    "code.google.com/p/go-uuid/uuid"
    "errors"
)


type Registry struct {
    sessions map[string]*Session
    lock sync.Mutex
}

func NewRegistry() *Registry {
    return &Registry{make(map[string]*Session), sync.Mutex{}}
}

func (r *Registry) CreateSession() *Session {
    r.lock.Lock()
    defer r.lock.Unlock()

    id := uuid.New()
    session := NewSession(id)
    go session.Run()
    r.sessions[id] = session
    return session
}

func (r *Registry) GetSession(sid string) *Session {
    return r.sessions[sid] // reading from maps is atomic
}

func (r *Registry) AttachPlayer(session_id string) (player string, err error) {
    r.lock.Lock()
    defer r.lock.Unlock()

    session, ok := r.sessions[session_id]
    if ! ok {
        return "", errors.New("Session not found")
    }
    player = uuid.New()
    err = session.AttachPlayer(player)
    if err != nil {
        return "", err
    }
    return player, nil
}
