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

import "log"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]struct{}

	// Inbound messages from the clients.
	broadcast chan string

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]struct{}),
	}
}

// Register a client with the hub.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// Broadcast to all clients in the hub.
func (h *Hub) Broadcast(s string) {
	h.broadcast <- s
}

func (h *Hub) run() {
	remove := func(c *Client) {
		c.Halt()
		delete(h.clients, c)
	}
	for {
		select {
		case client, ok := <-h.register:
			if !ok {
				log.Fatalln("Register channel closed unexpectedly")
			}
			h.clients[client] = struct{}{}
		case client, ok := <-h.unregister:
			if !ok {
				log.Fatalln("Register channel closed unexpectedly")
			}
			if _, ok := h.clients[client]; ok {
				// The client closed the connection
				remove(client)
			}
		case message, ok := <-h.broadcast:
			if !ok {
				log.Fatalln("Broadcaste channel closed unexpectedly")
			}
			for client := range h.clients {
				if !client.Send(message) {
					// The client send buffer was full.
					remove(client)
				}
			}
		}
	}
}
