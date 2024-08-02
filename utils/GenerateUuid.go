package utils

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func GenerateUUid() (string, error) {
	// Generate a new UUID.
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}

func GenerateClassroomId() *string {
	// Seed the random number generator to get different results each time
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 10000 and 99999
	min := 10000
	max := 99999
	randomNumber := rand.Intn(max-min+1) + min

	// Convert the random number to a string
	randomString := strconv.Itoa(randomNumber)

	return &randomString
}
