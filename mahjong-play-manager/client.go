// Copyright (c) 2013 The Gorilla WebSocket Authors. All rights reserved.
// https://github.com/gorilla/websocket/blob/master/LICENSE

package main

import (
	"log"
	"net/http"
	"time"
	"encoding/json"
	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
	conn *websocket.Conn
	send chan []byte
	playerId int
}

type Operator struct {
	Operation string
	Target int
}

func (c *Client) readPump(m *MahjongPlayManager) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		operator := c.parseOperator(message)
		sendBroadCast := true
		switch {
		case operator.isStart():
			m.InitRound()

			m.SendMessageStart()
		case operator.isDiscard():
			discardedTile := m.DiscardTile(operator.Target)
			canRon := m.CheckPinfuAndSetRon(discardedTile)
			if canRon {
				m.SendMessageDiscard(m.playerIdInTurn)
				m.SendMessageDiscardOther(m.playerIdInTurn, discardedTile)
			} else {
				if m.CanDistributeTile() {
					playerIdInTurnBefore := m.RotatePlayer()
					m.DistributeTile()

					m.SendMessageDiscard(playerIdInTurnBefore)
					m.SendMessageDiscardOther(playerIdInTurnBefore, discardedTile)
					m.SendMessageDrawn(discardedTile)
				} else {
					m.WaitNextMessage()

					m.SendMessageDrawnRound()
				}
			}
		case operator.isRon():
			ronInfo := m.CalculateRonInfo(c.playerId)
			m.UpdatePlayersPoint(ronInfo)
			m.SetFirstPinfuOrder(c.playerId)
			m.DealerWin(c.playerId)
			m.WaitNextMessage()

			m.SendMessageRon(ronInfo)
		case operator.isSkip():
			m.RotatePlayer()
			if m.CanDistributeTile() {
				m.DistributeTile()

				m.SendMessageSkip()
				m.SendMessageDrawn(tileIdNone)
			} else {
				m.WaitNextMessage()

				m.SendMessageDrawnRound()
			}
		case operator.isNext():
			f := func() {
				if m.continueGame() {
					if m.IsDealerWin() {
						m.NextSubRound()
					} else {
						m.RotateRound()
						m.RotatePlayerWind()
						m.ResetSubRound()
					}
					m.InitRound()

					m.SendMessageNext()
				} else {
					result := m.CalculateResult()

					m.SendMessageResult(result)
				}
			}
			sendBroadCast = m.TriggerNextMessage(f)
		case operator.isResult():
			result := m.CalculateResult()

			m.SendMessageResult(result)
		default:
			log.Println("not operated")
			continue
		}
		if sendBroadCast {
			c.hub.broadcast <- message
		}
	}
}

func (c *Client) parseOperator(message []byte) *Operator {
	operator := Operator{"", tileIdNone}
	err := json.Unmarshal(message, &operator)
	if err == nil {
		log.Println(operator)
	}
	return &operator
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (o *Operator) isStart() bool {
	return o.Operation == "start"
}

func (o *Operator) isDiscard() bool {
	return o.Operation == "discard"
}

func (o *Operator) isRon() bool {
	return o.Operation == "ron"
}

func (o *Operator) isSkip() bool {
	return o.Operation == "skip"
}

func (o *Operator) isNext() bool {
	return o.Operation == "next"
}

func (o *Operator) isResult() bool {
	return o.Operation == "result"
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("serveWs")
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), playerId: hub.mahjongPlayManager.newPlayerNumber()}
	client.hub.register <- client
	if hub.mahjongPlayManager.isReady() {
		hub.mahjongPlayManager.InitRound()
		hub.mahjongPlayManager.SendMessageStart()
		hub.broadcast <- []byte{}
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump(hub.mahjongPlayManager)
}
