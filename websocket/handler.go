package websocket

import (
	"Go-chat/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var rooms = make(map[int]*Room) // Room ID -> Room 매핑

// WebSocket 핸들러 (방 선택 가능)
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// WebSocket 연결 업그레이드
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("웹소켓 업그레이드 실패: %v", err)
		return
	}
	defer conn.Close()

	// 클라이언트가 선택한 방 ID 가져오기
	roomIDStr := r.URL.Query().Get("room_id")
	if roomIDStr == "" {
		log.Printf("방 ID가 제공되지 않았습니다.")
		return
	}
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil || rooms[roomID] == nil {
		log.Printf("유효하지 않은 방 ID: %s", roomIDStr)
		return
	}

	// 고유 사용자 생성
	user := &models.User{
		UserID:   generateUniqueID(),
		UserName: "User_" + generateUniqueID(),
	}

	// 방에 사용자 추가
	room := rooms[roomID]
	room.AddParticipant(conn, user)
	defer room.RemoveParticipant(conn)

	// 메시지 처리 루프
	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("메시지 읽기 실패: %v", err)
			break
		}
		room.Broadcast <- &msg
	}
}

// 방 생성 핸들러
func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("name")
	if roomName == "" {
		http.Error(w, "방 이름을 제공해야 합니다.", http.StatusBadRequest)
		return
	}

	roomID := len(rooms) + 1 // 방 ID 생성
	if _, exists := rooms[roomID]; exists {
		http.Error(w, "이미 존재하는 방입니다.", http.StatusConflict)
		return
	}

	rooms[roomID] = NewRoom(roomID, roomName)
	log.Printf("새 채팅방 생성: %s (ID: %d)", roomName, roomID)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"room_id":   roomID,
		"room_name": roomName,
	}
	json.NewEncoder(w).Encode(response)
}

// 방 목록 반환 핸들러
func ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	type RoomInfo struct {
		RoomID   int    `json:"room_id"`
		RoomName string `json:"room_name"`
	}

	var roomList []RoomInfo
	for _, room := range rooms {
		roomList = append(roomList, RoomInfo{
			RoomID:   room.RoomID,
			RoomName: room.RoomName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roomList); err != nil {
		http.Error(w, "방 목록 반환 실패", http.StatusInternalServerError)
	}
}

func DeleteRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteRoomHandler 호출됨") // 디버깅 로그 추가

	roomIDStr := r.URL.Query().Get("room_id")
	if roomIDStr == "" {
		log.Println("방 ID를 제공하지 않음")
		http.Error(w, "방 ID를 제공해야 합니다.", http.StatusBadRequest)
		return
	}

	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		log.Printf("유효하지 않은 방 ID: %s", roomIDStr)
		http.Error(w, "유효하지 않은 방 ID입니다.", http.StatusBadRequest)
		return
	}

	log.Printf("삭제 요청된 Room ID: %d", roomID)

	room, exists := rooms[roomID]
	if !exists {
		log.Printf("방 ID %d가 존재하지 않음", roomID)
		http.Error(w, "해당 방이 존재하지 않습니다.", http.StatusNotFound)
		return
	}

	room.CloseAllConnections()
	delete(rooms, roomID)

	log.Printf("채팅방 삭제 성공: %d", roomID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("채팅방 %d 삭제 성공", roomID)))
}

// 고유 ID 생성
func generateUniqueID() string {
	return time.Now().Format("20060102150405.000")
}
