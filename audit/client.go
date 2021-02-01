// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package audit

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

func init() {
	hub = newHub()
	go hub.run()
}

const (
	// Time allowed to write a message to the peer.
	// writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	// pongWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var hub *Hub

type Client struct {
	hub *Hub

	conn *websocket.Conn

	cond urlParam

	send chan audit.EventList
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		err := c.conn.ReadJSON(&c.cond)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				klog.Info(err)
			}
			break
		}

		query := queryBuilder(c.cond)
		eventList, _ := get(query)

		respMsg, err := json.Marshal(eventList)

		c.conn.WriteMessage(websocket.TextMessage, respMsg)
		if err != nil {
			klog.Error(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			}
			break
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			t, err := json.Marshal(message)
			if err != nil {
				klog.Info(err)
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, t)
		}
	}
}

func ServeWss(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan audit.EventList, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
