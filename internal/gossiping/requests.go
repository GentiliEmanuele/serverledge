package gossiping

import (
	"time"

	"github.com/serverledge-faas/serverledge/internal/function"
)

// find return true if there is an element with the same timestamp and the same function name
func (requests *Requests) find(f function.Function, timestamp time.Time) bool {
	for _, req := range *requests {
		if req.Timestamp.Equal(timestamp) && req.F.Name == f.Name {
			return true
		}
	}

	return false
}

func (requests *Requests) add(f function.Function, timestamp time.Time) {
	*requests = append(*requests, Request{
		Timestamp: timestamp,
		F:         f,
	})
}
