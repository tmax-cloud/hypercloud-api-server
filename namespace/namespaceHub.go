package namespace

import (
	"encoding/json"
	"strings"

	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type Hub struct {
	clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client
}

type NamespaceEvent struct {
	Type   string           `json:"type"`
	Object corev1.Namespace `json:"object"`
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			// Register client
			h.clients[client] = true

			// Send namespace list for the first time when a connection is created
			// nsList, err := GetNSList(client.cond.UserId, client.cond.LabelSelector, client.cond.UserGroup, client.cond.Limit)
			// if err != nil {
			// 	resp := "Failed to get namespace list"
			// 	client.send <- []byte(resp)
			// }
			// client.send <- nsList
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case ns_body_byte := <-h.broadcast:
			body := strings.Replace(string(ns_body_byte), `\"`, `"`, -1) // Remove `\` character
			body = body[1 : len(body)-1]                                 // Trim both side of `"` character
			body = strings.Replace(body, `\\`, `\`, -1)                  // remove duplicate '\' character to only one character
			var data NamespaceEvent
			if err := json.Unmarshal([]byte(body), &data); err != nil {
				klog.Error(err, " Failed to unmarshal namespace body")
				continue
			}
			event := data.Type
			ns := data.Object.Name
			klog.Infoln("namespace [" + ns + "] " + event + " detected")

			for client := range h.clients {
				isAccessible, err := k8sApiCaller.IsAccessibleNS(ns, client.cond.UserId, client.cond.LabelSelector, client.cond.UserGroup)
				if err != nil {
					resp := "Failed to check accessibility to namespace [" + ns + "]"
					client.send <- []byte(resp)
				}
				if !isAccessible {
					continue
				}
				select {
				case client.send <- []byte(body):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
