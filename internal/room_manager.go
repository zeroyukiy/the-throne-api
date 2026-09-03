package internal

import "sync"

type RoomRepository interface {
	CreateRoom(roomId string) *Room
	GetRoom(roomId string) (*Room, bool)
	RemoveRoom(roomId string) bool
	ListRooms() []*Room
}

type RoomManager struct {
	rooms map[string]*Room
	sync.RWMutex
}

func NewRoomManager() RoomRepository {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (m *RoomManager) CreateRoom(room_id string) *Room {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	if room := m.rooms[room_id]; room == nil {
		room := NewRoom(room_id, m)
		m.rooms[room_id] = room
		go room.Run()
		return room
	}
	return nil
}

func (m *RoomManager) GetRoom(room_id string) (*Room, bool) {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	if room := m.rooms[room_id]; room != nil {
		return room, true
	}
	return nil, false
}

func (m *RoomManager) RemoveRoom(room_id string) bool {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	if room := m.rooms[room_id]; room != nil {
		delete(m.rooms, room_id)
		return true
	}
	return false
}

func (m *RoomManager) ListRooms() []*Room {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	list := []*Room{}
	for _, r := range m.rooms {
		list = append(list, r)
	}
	return list
}
