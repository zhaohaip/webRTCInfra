package app

import (
	"sync"
)

type ClientManager struct {
	Clients sync.Map
}

func NewClientManager() *ClientManager {
	return &ClientManager{}
}

func (cm *ClientManager) AddClient(client *Client) {
	cm.Clients.Store(client.UserID, client)
}

func (cm *ClientManager) GetClient(userId string) (*Client, bool) {
	c, ok := cm.Clients.Load(userId)
	if !ok {
		return nil, false
	}
	return c.(*Client), true
}

func (cm *ClientManager) RemoveClient(userId string) {
	cm.Clients.Delete(userId)
}

func (cm *ClientManager) ListClients() []*Client {
	var clients []*Client
	cm.Clients.Range(func(key, value interface{}) bool {
		clients = append(clients, value.(*Client))
		return true
	})
	return clients
}
