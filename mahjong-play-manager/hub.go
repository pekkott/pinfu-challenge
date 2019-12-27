// Copyright (c) 2013 The Gorilla WebSocket Authors. All rights reserved.
// https://github.com/gorilla/websocket/blob/master/LICENSE

package main
import "log"

type Hub struct {
	clients map[*Client]bool
	broadcast chan []byte
	register chan *Client
	unregister chan *Client
	mahjongPlayManager *MahjongPlayManager
}

func newHub(m *MahjongPlayManager) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		mahjongPlayManager:    m,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			var _ = message
			for client := range h.clients {
				log.Printf("client playerId:%d", client.playerId)
				select {
				case client.send <- h.mahjongPlayManager.sendMessages[client.playerId].ToBytes():
                                log.Printf("sendMessage:%s", h.mahjongPlayManager.sendMessages[client.playerId].ToBytes())
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
