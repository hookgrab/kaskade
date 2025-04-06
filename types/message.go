package types

import (
	hgp "hg.atrin.dev/proto/gen/go/proto"
)

type Message struct {
	Req *hgp.Request
	Uid *string
	HUid int64
}
