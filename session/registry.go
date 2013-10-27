package session

import (
    "sync"
    "errors"
)


type Registry struct {
    sessions map[string]*Session
    lock sync.Mutex
}

func NewRegistry() *Registry {
    return &Registry{make(map[string]*Session), sync.Mutex{}}
}

func (r *Registry) CreateSession(id string) (*Session, error) {
    r.lock.Lock()
    defer r.lock.Unlock()

    if id == "" {
        return nil, errors.New("Empty session id")
    }

    _, ok := r.sessions[id]
    if ok {
        // already have one
        return nil, errors.New("Session " + id + " already exists")
    }

    session := NewSession(id)
    go session.Run()
    r.sessions[id] = session
    return session, nil
}

func (r *Registry) GetSession(sid string) *Session {
    return r.sessions[sid] // reading from maps is atomic
}

func (r *Registry) AttachPlayer(sessionId, playerId string) (player, glyph string, err error) {
    r.lock.Lock()
    defer r.lock.Unlock()

    session, ok := r.sessions[sessionId]
    if !ok {
        return "", "", errors.New("Session not found")
    }
    glyph, err = session.AttachPlayer(playerId)
    if err != nil {
        return "", "", err
    }
    return playerId, glyph, nil
}
