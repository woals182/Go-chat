package websocket

import (
	"Go-chat/models"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Room 구조체
type Room struct {
	RoomID       int
	RoomName     string
	Participants map[*websocket.Conn]*models.User // Connection -> User 매핑
	Broadcast    chan *models.Message             // 메시지 브로드캐스트 채널
	mu           sync.Mutex                       // 동시성 제어
}

// Room 생성
func NewRoom(roomID int, roomName string) *Room {
	room := &Room{
		RoomID:       roomID,
		RoomName:     roomName,
		Participants: make(map[*websocket.Conn]*models.User),
		Broadcast:    make(chan *models.Message),
	}

	// 메시지 브로드캐스트 시작
	go room.StartBroadcast()

	return room
}

func (r *Room) CloseAllConnections() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("Room ID %d: 모든 연결 닫기 시작", r.RoomID)

	for conn, user := range r.Participants {
		log.Printf("참가자 연결 닫기: %s (%s)", user.UserName, user.UserID)
		conn.Close()
		delete(r.Participants, conn)
	}

	log.Printf("Room ID %d: 모든 연결 닫기 완료", r.RoomID)
}

// 참여자 추가
func (r *Room) AddParticipant(conn *websocket.Conn, user *models.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Participants[conn] = user

	// 참여 알림 메시지 생성
	joinMessage := &models.Message{
		RoomID:    r.RoomID,
		UserID:    user.UserID,
		Content:   user.UserName + "님이 입장하셨습니다.",
		Timestamp: "",
	}
	r.Broadcast <- joinMessage
	log.Printf("%s님이 방에 참여했습니다.", user.UserID)
}

// 참여자 제거
func (r *Room) RemoveParticipant(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, exists := r.Participants[conn]
	if exists {
		// 퇴장 알림 메시지 생성
		leaveMessage := &models.Message{
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

// 메시지 브로드캐스트
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
