package models

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// ✅ 사용자(User) 구조체
type User struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// ✅ 참여자(Participant) 구조체 (JSON 직렬화 가능)
type Participant struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// ✅ 메시지(Message) 구조체
type Message struct {
	MessageID string `json:"message_id"`
	RoomID    int    `json:"room_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// ✅ 채팅방(Room) 구조체 (WebSocket 기능 유지)
type Room struct {
	RoomID       int                       `json:"room_id"`
	RoomName     string                    `json:"room_name"`
	CreaterID    string                    `json:"creater_id"`
	Participants map[*websocket.Conn]*User // ✅ WebSocket 연결된 사용자 매핑
	Broadcast    chan *Message             // ✅ 메시지 브로드캐스트 채널
	mu           sync.Mutex                // ✅ 동시성 제어
}

// ✅ JSON 변환 가능한 채팅방 구조체
type SerializableRoom struct {
	RoomID       int           `json:"room_id"`
	RoomName     string        `json:"room_name"`
	CreaterID    string        `json:"creater_id"`
	Participants []Participant `json:"participants"` // ✅ JSON 변환 가능
}

// ✅ 방을 JSON 변환 가능하도록 변환하는 메서드 추가
func (r *Room) ToSerializable() SerializableRoom {
	var participants []Participant
	for _, user := range r.Participants { // ✅ `map[*websocket.Conn]*User` → `[]Participant`
		participants = append(participants, Participant{
			UserID:   user.UserID,
			UserName: user.UserName,
		})
	}

	if participants == nil {
		participants = []Participant{}
	}

	return SerializableRoom{
		RoomID:       r.RoomID,
		RoomName:     r.RoomName,
		CreaterID:    r.CreaterID,
		Participants: participants,
	}
}

// ✅ 방 생성 함수
func NewRoom(roomID int, roomName, createrID string) *Room {
	room := &Room{
		RoomID:       roomID,
		RoomName:     roomName,
		CreaterID:    createrID,
		Participants: make(map[*websocket.Conn]*User),
		Broadcast:    make(chan *Message),
	}

	// ✅ 메시지 브로드캐스트 시작
	go room.StartBroadcast()

	return room
}

// ✅ 방에 있는 모든 WebSocket 연결 닫기
func (r *Room) CloseAllConnections() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Room ID %d: 모든 연결 닫기 시작", r.RoomID)

	for conn, user := range r.Participants {
		if conn != nil { // ✅ conn이 nil인지 확인
			log.Printf("참가자 연결 닫기: %s (%s)", user.UserName, user.UserID)
			err := conn.Close()
			if err != nil {
				log.Printf("🚨 WebSocket 연결 닫기 오류: %v", err)
			}
		}
		delete(r.Participants, conn) // ✅ 참가자 삭제
	}

	log.Printf("Room ID %d: 모든 연결 닫기 완료", r.RoomID)
}

// ✅ 참여자 추가
func (r *Room) AddParticipant(conn *websocket.Conn, user *User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Participants[conn] = user

	// ✅ 참여 알림 메시지 생성
	joinMessage := &Message{
		RoomID:    r.RoomID,
		UserID:    user.UserID,
		Content:   user.UserName + "님이 입장하셨습니다.",
		Timestamp: "",
	}
	r.Broadcast <- joinMessage
	log.Printf("%s님이 방에 참여했습니다.", user.UserID)
}

// ✅ 참여자 제거
func (r *Room) RemoveParticipant(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.Participants[conn]
	if exists {
		// ✅ 퇴장 알림 메시지 생성
		leaveMessage := &Message{
			RoomID:    r.RoomID,
			UserID:    user.UserID,
			Content:   user.UserName + "님이 퇴장하셨습니다.",
			Timestamp: "",
		}
		r.Broadcast <- leaveMessage
		log.Printf("%s님이 방을 나갔습니다.", user.UserID)
		delete(r.Participants, conn)
	}
}

// ✅ 메시지 브로드캐스트
func (r *Room) StartBroadcast() {
	for {
		message := <-r.Broadcast
		r.mu.Lock()
		for conn := range r.Participants {
			if err := conn.WriteJSON(message); err != nil {
				log.Printf("메시지 전송 실패: %v", err)
				conn.Close()
				delete(r.Participants, conn)
			}
		}
		r.mu.Unlock()
	}
}
