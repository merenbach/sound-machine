// Copyright 2018 Andrew Merenbach
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"io"

	"github.com/gin-gonic/gin"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The server-sent event name.
	event string

	// Buffered channel of outbound messages.
	send chan string
}

// Send to the client, returning whether the operation was successful.
func (c *Client) Send(s string) bool {
	select {
	case c.send <- s:
		return true
	default:
		return false
	}
}

// registerClient handles SSE intitiation requests from the peer.
func registerClient(hub *Hub, c *gin.Context) {
	client := &Client{event: "message", send: make(chan string, 256)}
	hub.Register(client)
	defer hub.Unregister(client)

	c.Stream(func(w io.Writer) bool {
		if message, ok := <-client.send; ok {
			c.SSEvent(client.event, message)
			return true
		}
		return false
	})
}
