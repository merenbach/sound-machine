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

// Portions copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// Time allowed to write a message to the peer.
	// writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// maxMessageSize = 512
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

// Halt further communications to this client by closing its send channel.
func (c *Client) Halt() {
	close(c.send)
}

func (c *Client) writePump(ctx *gin.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// c.conn.Close()
	}()
	ctx.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-c.send:
			// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return false
			}

			/*w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return false
			}*/
			ctx.SSEvent(c.event, message)

			// Add queued chat messages to the current websocket message.
			/*n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				ctx.SSEvent(c.event, <-c.send)
			}*/

			/*if err := w.Close(); err != nil {
				return false
			}*/
		case <-ticker.C:
			log.Println("Keeping alive")
			ctx.SSEvent("", "Keep alive")
			//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			/*if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return false
			}*/
		}
		return true
	})
}

func newClient(e string) *Client {
	return &Client{
		event: e,
		send:  make(chan string, 256),
	}
}
