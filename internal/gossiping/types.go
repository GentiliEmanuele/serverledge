package gossiping

import (
	"time"

	"github.com/serverledge-faas/serverledge/internal/function"
	"github.com/serverledge-faas/serverledge/internal/registration"
)

type Requests []Request

type Request struct {
	F         function.Function
	Timestamp time.Time
}

type NodeList []registration.NodeRegistration
