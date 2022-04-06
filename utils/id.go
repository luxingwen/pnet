package utils

import (
	"github.com/google/uuid"
)

func GenId() []byte {
	return []byte(uuid.New().String())
}
