package websocket

import (
	"sync"
)

type Manager struct {
	Clients sync.Map
}

func NewManager() *Manager {
	return &Manager{}
}

func (cm *Manager) AddClient(client *Connection) {
	cm.Clients.Store(client.UserID, client)
}

func (cm *Manager) GetClient(userId string) (*Connection, bool) {
	c, ok := cm.Clients.Load(userId)
	if !ok {
		return nil, false
	}
	return c.(*Connection), true
}

func (cm *Manager) RemoveClient(userId string) {
	cm.Clients.Delete(userId)
}

func (cm *Manager) ListClients() []*Connection {
	var clients []*Connection
	cm.Clients.Range(func(key, value interface{}) bool {
		clients = append(clients, value.(*Connection))
		return true
	})
	return clients
}
