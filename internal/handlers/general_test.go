package handlers

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleteLostUserRequests(t *testing.T) {
	var mu sync.RWMutex

	timeNow := time.Now()

	cases := []struct {
		name               string
		deleteAfterSeconds int
		userRequests       map[int]Checker
		expected           map[int]Checker
	}{
		{
			name:               "positive_deleted",
			deleteAfterSeconds: 5,
			userRequests: map[int]Checker{
				1: {
					Added: timeNow.Add(time.Second * -6),
				},
			},
			expected: map[int]Checker{},
		},
		{
			name:               "positive_not_deleted",
			deleteAfterSeconds: 5,
			userRequests: map[int]Checker{
				1: {
					Added: timeNow.Add(time.Second * -4),
				},
			},
			expected: map[int]Checker{
				1: {
					Added: timeNow.Add(time.Second * -4),
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			deleteLostUserRequests(&mu, tt.userRequests, tt.deleteAfterSeconds)

			assert.Equal(t, tt.expected, tt.userRequests)
		})
	}
}
