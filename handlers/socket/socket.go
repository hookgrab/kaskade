package socket

import (
	"slices"

	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/contrib/websocket"
	"hg.atrin.dev/kaskade/types"
)

func HandleWsConn(workers chan struct{}, c *websocket.Conn, cl *types.Client, uid string, huid int64) error {
	workers <- struct{}{}
	defer func() { 
		log.Info("Closing connection", uid, huid)
		cl.Conns = slices.DeleteFunc(cl.Conns, func(x *types.Connection) bool { return c == x.Ws })
		c.Close()
		<- workers 
	}()
	log.Debug("Handler", c,)
	for {
		msgt, _, err := c.ReadMessage()

		if err != nil {
			log.Error("Failed to read message", c, err)
			return err
		}
		if msgt == websocket.TextMessage {
		} else if msgt == websocket.BinaryMessage {
		}
	}
	return nil
}
