// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

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
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Halt()
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if !client.Send(message) {
					client.Halt()
					delete(h.clients, client)
				}
			}
		}
	}
}
