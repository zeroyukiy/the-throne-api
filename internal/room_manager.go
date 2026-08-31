package internal

import (
	"fmt"
)

type RoomRepository interface {
	CreateRoom(roomId string) *Room
	GetRoom(roomId string) (*Room, bool)
	RemoveRoom(roomId string) bool
	ListRooms() []*Room
}

type RoomManager struct {
	rooms map[string]*Room
}

func NewRoomManager(rooms []string) RoomRepository {
	roomManager := &RoomManager{
		rooms: make(map[string]*Room),
	}

	for _, name := range rooms {
		if room := roomManager.CreateRoom(name); room != nil {
			fmt.Printf("%s room created\n", name)
		}
	}

	return roomManager
}

func (m *RoomManager) CreateRoom(room_id string) *Room {
	if room := m.rooms[room_id]; room == nil {
		room := NewRoom(room_id, m)
		m.rooms[room_id] = room
		go room.Run()
		return room
	}
	return nil
}

func (m *RoomManager) GetRoom(room_id string) (*Room, bool) {
	if room := m.rooms[room_id]; room != nil {
		return room, true
	}
	return nil, false
}

func (m *RoomManager) RemoveRoom(room_id string) bool {
	if room := m.rooms[room_id]; room != nil {
		delete(m.rooms, room_id)
		return true
	}
	return false
}

func (m *RoomManager) ListRooms() []*Room {
	list := []*Room{}
	for _, r := range m.rooms {
		list = append(list, r)
	}
	return list
}
