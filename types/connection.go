package types

import (
	"github.com/gofiber/contrib/websocket"
	"sync"
)

type Connection struct {
	Ws *websocket.Conn
	IsClosing bool
}

type Client struct {
	Uid string
	Mu sync.Mutex

	Conns []*Connection
}
