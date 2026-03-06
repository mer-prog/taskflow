package ws

import (
	"encoding/json"
	"log"
	"sync"
)

type HubManager struct {
	mu   sync.RWMutex
	hubs map[string]*Hub
}

func NewHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*Hub),
	}
}

func (m *HubManager) GetOrCreateHub(boardID string) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if h, ok := m.hubs[boardID]; ok {
		return h
	}

	h := newHub(boardID, m)
	m.hubs[boardID] = h
	go h.run()
	return h
}

func (m *HubManager) removeHub(boardID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.hubs, boardID)
}

// Broadcast sends a message to all clients connected to the given board.
func (m *HubManager) Broadcast(boardID string, msg WSMessage) {
	m.mu.RLock()
	h, ok := m.hubs[boardID]
	m.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: failed to marshal message: %v", err)
		return
	}

	h.broadcast <- data
}

type Hub struct {
	boardID    string
	manager    *HubManager
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub(boardID string, manager *HubManager) *Hub {
	return &Hub{
		boardID:    boardID,
		manager:    manager,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.broadcastMemberEvent("member:joined", client.userID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.broadcastMemberEvent("member:left", client.userID)

				if len(h.clients) == 0 {
					h.manager.removeHub(h.boardID)
					return
				}
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (h *Hub) broadcastMemberEvent(eventType, userID string) {
	payload, _ := json.Marshal(map[string]string{"user_id": userID})
	msg := WSMessage{
		Type:    eventType,
		Payload: payload,
		UserID:  userID,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			delete(h.clients, client)
			close(client.send)
		}
	}
}
