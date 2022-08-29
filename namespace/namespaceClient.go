// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package namespace

import (
	"github.com/gorilla/websocket"
	"k8s.io/klog"
)

func init() {
	hub = newHub()
	go hub.run()
}

type urlParam struct {
	UserId        string   `json:"userId"`
	Limit         string   `json:"limit"`
	LabelSelector string   `json:"labelSelector"`
	UserGroup     []string `json:"userGroup"`
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

	send chan []byte
}

// Whenever a client send a message to server,
// the server will give namespace list.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				klog.V(3).Info(err)
			}
			break
		}
		nsList, err := GetNSList(c.cond.UserId, c.cond.LabelSelector, c.cond.UserGroup, c.cond.Limit)
		if err != nil {
			resp := "Failed to get namespace list"
			c.conn.WriteMessage(websocket.TextMessage, []byte(resp))
		}
		err = c.conn.WriteMessage(websocket.TextMessage, nsList)
		if err != nil {
			klog.V(1).Info(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			}
			break
		}
	}
}

// When Hypercloud-Single-Operator detects namespace create/delete events,
// it will call hypercloud-api-server API and thus writePump() will run, which sends namespace list to all clients.
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case nsList, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, nsList)
			if err != nil {
				klog.V(1).Infoln(err)
			}
		}
	}
}
