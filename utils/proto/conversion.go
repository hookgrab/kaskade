package proto

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	hgp "hg.atrin.dev/proto/gen/go/proto"
)

func ConvertMethod(c *fiber.Ctx) (hgp.METHOD, error) {
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

func GenRequestObj(c *fiber.Ctx, uid *string) (*hgp.Request, error) {
	method, err := ConvertMethod(c)
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
