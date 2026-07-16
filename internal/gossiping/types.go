package gossiping

import (
	"time"

	"github.com/serverledge-faas/serverledge/internal/function"
)

type Requests []Request

type Request struct {
	F         function.Function
	Timestamp time.Time
}
