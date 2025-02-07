package utils

import (
	"encoding/base64"
	"github.com/google/uuid"
	"strings"
)

func GenerateRandomId() string {
	id := uuid.New()
	idString := base64.RawStdEncoding.EncodeToString(id[:])[:8]
	if strings.Contains(idString, "/") {
		idString = strings.Replace(idString, "/", "A", -1)
		return idString
	}
	return idString
}
