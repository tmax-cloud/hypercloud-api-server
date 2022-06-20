package namespace

type Hub struct {
	clients map[*Client]bool

	broadcast chan bool

	register chan *Client

	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan bool),
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
			nsList, err := GetNSList(client.cond.UserId, client.cond.LabelSelector, client.cond.UserGroup, client.cond.Limit)
			if err != nil {
				resp := "Failed to get namespace list"
				client.send <- []byte(resp)
			}
			client.send <- nsList
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case listChanged := <-h.broadcast:
			if listChanged == false {
				break
			}
			for client := range h.clients {
				nsList, err := GetNSList(client.cond.UserId, client.cond.LabelSelector, client.cond.UserGroup, client.cond.Limit)
				if err != nil {
					resp := "Failed to get namespace list"
					client.send <- []byte(resp)
				}
				select {
				case client.send <- nsList:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
