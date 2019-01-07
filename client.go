// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"

	"github.com/gin-gonic/gin"
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	context *gin.Context

	// Buffered channel of outbound messages.
	send chan []byte
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	c.context.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-c.send:
			if !ok {
				return false
			}
			c.context.SSEvent("message", string(message))

			// TODO: is this necessary with SSE?
			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				c.context.SSEvent("message", string(<-c.send))
			}
			return true
		}
	})
}

// registerClient handles SSE intitiation requests from the peer.
func registerClient(hub *Hub, c *gin.Context) {
	client := &Client{hub: hub, context: c, send: make(chan []byte, 256)}
	client.hub.register <- client
	defer func() {
		client.hub.unregister <- client
	}()

	client.writePump()
}
