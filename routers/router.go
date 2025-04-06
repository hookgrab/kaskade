package routers

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	handlers "hg.atrin.dev/kaskade/handlers/socket"
	utils "hg.atrin.dev/kaskade/utils"
	protoutils "hg.atrin.dev/kaskade/utils/proto"
	"hg.atrin.dev/kaskade/types"
)

var (
	workers chan struct{} = nil;
)

func SetWorkers(wrs chan struct{}) {
	workers = wrs
}

func handleListenerReq(c *fiber.Ctx) error {
	uid := c.Params("uid")
	huid, err := utils.HashUID(&uid)

	if err != nil {
		return c.Status(400).SendString("Invliad UID")
	}

	c.Locals("HUID", huid)

	log.Println("Sending message", uid, huid)

	req, err := protoutils.GenRequestObj(c, &uid)
	
	if err != nil {
		log.Fatalln("Failed to convert request", err)
		c.Status(400).SendString("An error has occured: Couldn't parse request")
		return err
	}

	msg := types.Message{
		Uid: &uid,
		HUid: huid,

		Req: req,
	}
	handlers.GetMessageTunnel() <- &msg

	return c.SendString("UID is " + uid)
}

func handleWebsocketReq(c *websocket.Conn) {
	uid := c.Params("uid")
	// huid, err := hashuid(&uid)
	huid, ok := c.Locals("HUID").(int64)
	log.Println("Websocket opened", uid, huid)
	if !ok {
		c.Close()
		return 
	}
	c.WriteMessage(websocket.TextMessage, []byte("Hello"))

	cls, ok := handlers.GetClients(huid)
	var cl *types.Client = nil
	for _, client := range cls {
		if client.Uid == uid {
			cl = client
			break
		}
	}

	err := c.WriteMessage(websocket.TextMessage, []byte("Hello"))
	if err != nil {
		log.Println("Failed to send message", err)

	}

	if ok && cl != nil {
		conn := &types.Connection{ Ws: c, IsClosing: false }
		cl.Conns = append(cl.Conns, conn)
	} else {
		conn := types.Connection{ Ws: c, IsClosing: false }
		conns := make([]*types.Connection, 1)
		conns[0] = &conn
		newclient := &types.Client{ Uid: uid, Conns: conns }
		if ok {
			cls = append(cls, newclient)
		} else {
			handlers.CreateClientHUid(huid)[0] = newclient
		}
	}

	log.Println("Clients updated", c)

	handlers.HandleWsConn(workers, c, cl, uid, huid)

}


func SetupRouters(app *fiber.App) {
	app.Use("/ws/:uid", func(c *fiber.Ctx) error {
		log.Println("Connection received")
		// if c.Get("host") == "localhost:4000" {
		// 	c.Locals("Host", "localhost:4000")
		// 	return c.Next()
		// }

		uid := c.Params("uid")
		huid, err := utils.HashUID(&uid)

		log.Println("UID: ", uid)

		if err != nil {
			return c.Status(400).SendString("Invliad UID")
		}

		c.Locals("HUID", huid)

		log.Println("Upgrading connection, ", huid)

		return c.Next()
		return c.Status(403).SendString("CORS ERROR")
	})

	app.Get("/ws/:uid", websocket.New(handleWebsocketReq))

	app.Get("/l/:uid", handleListenerReq)
}
