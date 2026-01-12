package idx

import (
	"github.com/google/uuid"
	"log"
)

func NewUUID() uuid.UUID {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Println("fail to generate UUID", err)
		return uuid.UUID{}
	}
	return id
}
