package socket

import (
	"hg.atrin.dev/kaskade/types"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
	"google.golang.org/protobuf/proto"
)

var (
	tunnel = make(chan *types.Message)
	clients = make(map[int64][]*types.Client)
)

func GetMessageTunnel() chan *types.Message {
	return tunnel
}

func GetClients(huid int64) ([]*types.Client, bool) {
	cls, ok := clients[huid]
	return cls, ok
}

func CreateClientHUid(huid int64) []*types.Client {
	clients[huid] = make([]*types.Client, 1)
	return clients[huid]
}

func inform_connections(cl *types.Client, msg *types.Message) error {
	cl.Mu.Lock()
	defer cl.Mu.Unlock()

	reqbts, err := proto.Marshal(msg.Req)

	if err != nil {
		return err
	}

	for _, cn := range cl.Conns {
		if cn.IsClosing {
			continue
		}

		cn.Ws.WriteMessage(websocket.TextMessage, reqbts)
	}
	return nil
}

func MessageLoop() {
	for {
		msg, ok := <-tunnel

		if !ok {
			return 
		}

		log.Info("Message recieved", *msg.Uid, msg.HUid)

		cls, ok := clients[msg.HUid]

		if !ok {
			continue
		}

		for _, c := range cls {
			if c.Uid != *msg.Uid {
				continue
			}
			err := inform_connections(c, msg)

			if err != nil {
				log.Error("Failed inform connections", err)
			}
		}
	}
}
