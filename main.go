package main

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	hgp "hg.atrin.dev/proto/gen/go/proto"
	"google.golang.org/protobuf/proto"
)

type Connection struct {
	ws *websocket.Conn
	isClosing bool

}

type Client struct {
	uid string
	mu sync.Mutex

	conns []*Connection
}

var (
	clients = make(map[int64][]*Client)
)

type Message struct {
	req *hgp.Request
	uid *string
	huid int64

}

var c byte = 'c'

const MOD int64 = 1e9 + 7
const BASE int64 = 67

var revmap = [128]int64{
	0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
	1,2,3,4,5,6,7,8,9,10,0,0,0,0,0,0,0,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,
	33,34,35,36,0,0,0,0,0,0,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,
	0,0,0,0,0,
}


func hashuid(s *string) (int64, error) {
	var h int64
	h = 0
	for _, c := range *s {
		ind := int(c)

		if ind >= 128 {
			return 0, errors.New("Invalid Character")
		}
		id := revmap[ind]

		h = (h * BASE + id) % MOD
	}

	return h, nil
}

var (
	tunnel = make(chan *Message)
)

func inform_connections(cl *Client, msg *Message) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	reqbts, err := proto.Marshal(msg.req)

	if err != nil {
		return err
	}

	for _, cn := range cl.conns {
		if cn.isClosing {
			continue
		}

		cn.ws.WriteMessage(websocket.TextMessage, reqbts)
	}
	return nil
}

func message_handler() {
	for {
		msg, ok := <-tunnel

		if !ok {
			return 
		}

		log.Println("Message recieved", *msg.uid, msg.huid)

		cls, ok := clients[msg.huid]

		if !ok {
			continue
		}

		for _, c := range cls {
			if c.uid != *msg.uid {
				continue
			}
			err := inform_connections(c, msg)

			if err != nil {
				log.Fatalln("Failed inform connections", err)
			}
		}
	}
}

func handleWsConn(workers chan struct{}, c *websocket.Conn, cl *Client, uid string, huid int64) error {
	workers <- struct{}{}
	defer func() { 
		log.Println("Closing connection", uid, huid)
		cl.conns = slices.DeleteFunc(cl.conns, func(x *Connection) bool { return c == x.ws })
		c.Close()
		<- workers 
	}()
	log.Println("Handler", c,)
	for {
		msgt, _, err := c.ReadMessage()

		if err != nil {
			log.Println("Failed to read message", c, err)
			return err
		}
		if msgt == websocket.TextMessage {
		} else if msgt == websocket.BinaryMessage {
		}
	}
	return nil
}

func convertMethod(c *fiber.Ctx) (hgp.METHOD, error) {
	switch c.Method() {
	case fiber.MethodGet:
			return hgp.METHOD_GET, nil
	case fiber.MethodPost:
			return hgp.METHOD_POST, nil
	case fiber.MethodPut:
			return hgp.METHOD_PUT, nil
	case fiber.MethodPatch:
			return hgp.METHOD_PATCH, nil
	case fiber.MethodDelete:
			return hgp.METHOD_DELETE, nil
	case fiber.MethodHead:
			return hgp.METHOD_HEAD, nil
	default:
		return hgp.METHOD_UNKOWN_METHOD, fmt.Errorf("Unknown Method: %s", c.Method())
	}
}

func genRequestObj(c *fiber.Ctx, uid *string) (*hgp.Request, error) {
	method, err := convertMethod(c)
	if err != nil {
		return nil, err
	}
	req := &hgp.Request{
		WebhookUid: *uid,
		Method: method,
		Url: c.OriginalURL(),

		Protocol: hgp.PROTOCOL_HTTP_1_1,

		Body: c.BodyRaw(),
	}
	return req, nil
}

func main() {
	workers := make(chan struct{}, 100)

	app := fiber.New()

	go message_handler()

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Use("/ws/:uid", func(c *fiber.Ctx) error {
		log.Println("Connection received")
		// if c.Get("host") == "localhost:4000" {
		// 	c.Locals("Host", "localhost:4000")
		// 	return c.Next()
		// }

		uid := c.Params("uid")
		huid, err := hashuid(&uid)

		log.Println("UID: ", uid)

		if err != nil {
			return c.Status(400).SendString("Invliad UID")
		}

		c.Locals("HUID", huid)

		log.Println("Upgrading connection, ", huid)

		return c.Next()
		return c.Status(403).SendString("CORS ERROR")
	})

	app.Get("/ws/:uid", websocket.New(func (c *websocket.Conn) {
		uid := c.Params("uid")
		// huid, err := hashuid(&uid)
		huid, ok := c.Locals("HUID").(int64)
		log.Println("Websocket opened", uid, huid)
		if !ok {
			c.Close()
			return 
		}
		c.WriteMessage(websocket.TextMessage, []byte("Hello"))

		cls, ok := clients[huid]
		var cl *Client = nil
		for _, client := range cls {
			if client.uid == uid {
				cl = client
				break
			}
		}

		err := c.WriteMessage(websocket.TextMessage, []byte("Hello"))
		if err != nil {
			log.Println("Failed to send message", err)
			
		}

		if ok && cl != nil {
			conn := &Connection{ ws: c, isClosing: false }
			cl.conns = append(cl.conns, conn)
		} else {
			conn := Connection{ ws: c, isClosing: false }
			conns := make([]*Connection, 1)
			conns[0] = &conn
			newclient := &Client{ uid: uid, conns: conns }
			if ok {
				cls = append(cls, newclient)
			} else {
				clients[huid] = make([]*Client, 1)
				clients[huid][0] = newclient
			}
		}

		log.Println("Clients updated", c)

		handleWsConn(workers, c, cl, uid, huid)

	}))

	app.Get("/l/:uid", func(c *fiber.Ctx) error {

		uid := c.Params("uid")
		huid, err := hashuid(&uid)

		if err != nil {
			return c.Status(400).SendString("Invliad UID")
		}

		c.Locals("HUID", huid)

		log.Println("Sending message", uid, huid)

		req, err := genRequestObj(c, &uid)
		
		if err != nil {
			log.Fatalln("Failed to convert request", err)
			return c.Status(400).SendString("An error has occured: Couldn't parse request")
		}

		msg := Message{
			uid: &uid,
			huid: huid,

			req: req,
		}
		tunnel <- &msg

		return c.SendString("UID is " + uid)
	})

	app.Listen(":4000")
}
