package audit

import (
	"strconv"

	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

type Hub struct {
	clients map[*Client]bool

	broadcast chan audit.EventList

	register chan *Client

	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan audit.EventList),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
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
		case eventList := <-h.broadcast:
			for client := range h.clients {
				message := filter(eventList, client.cond)
				if len(message.Items) != 0 {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

func filter(eventList audit.EventList, cond urlParam) audit.EventList {
	out := audit.EventList{}
	out.Kind = "EventList"
	out.APIVersion = "audit.k8s.io/v1"
	code64, _ := strconv.ParseInt(cond.Code, 10, 32)

	klog.V(3).Info("NS cond:", cond.Namespace)
	klog.V(3).Info("RS cond:", cond.Resource)
	klog.V(3).Info("Code cond:", cond.Code)

	for _, data := range eventList.Items {
		ns := cond.Namespace == "" || (data.ObjectRef.Namespace == cond.Namespace)
		rs := cond.Resource == "" || (data.ObjectRef.Resource == cond.Resource)
		code := cond.Code == "" || ((data.ResponseStatus.Code/100)*100 == int32(code64))

		klog.V(3).Info("data.ObjectRef.Namespace:", data.ObjectRef.Namespace)
		klog.V(3).Info("data.ObjectRef.Resource:", data.ObjectRef.Resource)
		klog.V(3).Info("data.ResponseStatus.Code:", data.ResponseStatus.Code)

		klog.V(3).Info("NS:", ns)
		klog.V(3).Info("RS:", rs)
		klog.V(3).Info("Code:", code)

		if ns && rs && code {
			out.Items = append(out.Items, data)
		}
	}
	return out
}
