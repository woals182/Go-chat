package websocket

import (
	"Go-chat/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var rooms = make(map[int]*models.Room) // ✅ `models.Room` 사용

// ✅ WebSocket 핸들러 (채팅방 연결)
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("웹소켓 업그레이드 실패: %v", err)
		return
	}
	defer conn.Close()

	roomIDStr := c.Query("room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil || rooms[roomID] == nil {
		log.Printf("유효하지 않은 방 ID: %s", roomIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 방 ID입니다."})
		return
	}

	user := &models.User{
		UserID:   generateUniqueID(),
		UserName: "User_" + generateUniqueID(),
	}

	room := rooms[roomID]
	room.AddParticipant(conn, user)
	defer room.RemoveParticipant(conn)

	for {
		var msg models.Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("메시지 읽기 실패: %v", err)
			break
		}
		room.Broadcast <- &msg
	}
}

// ✅ 방 생성 API (`POST /create-room`)
func CreateRoomHandler(c *gin.Context) {
	var request struct {
		RoomName  string `json:"room_name" binding:"required"`
		CreaterID string `json:"creater_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "방 이름과 생성자 ID를 제공해야 합니다."})
		return
	}

	roomID := len(rooms) + 1
	newRoom := models.NewRoom(roomID, request.RoomName, request.CreaterID)

	// ✅ 생성자를 첫 번째 참가자로 추가 (방장 역할)
	newRoom.Participants[&websocket.Conn{}] = &models.User{
		UserID:   request.CreaterID,
		UserName: "방장",
	}

	rooms[roomID] = newRoom

	log.Printf("새 채팅방 생성: %s (ID: %d, 생성자: %s)", request.RoomName, roomID, request.CreaterID)

	// ✅ 생성된 방 정보 반환
	c.JSON(http.StatusOK, gin.H{
		"room_id":    newRoom.RoomID,
		"room_name":  newRoom.RoomName,
		"creater_id": newRoom.CreaterID,
		"participants": []models.Participant{
			{
				UserID:   request.CreaterID,
				UserName: "방장",
			},
		},
	})
}

// ✅ 방 목록 조회 API (`GET /list-rooms`)
func ListRoomsHandler(c *gin.Context) {
	var roomList []models.SerializableRoom // ✅ JSON 변환 가능한 구조체 사용

	for _, room := range rooms {
		roomList = append(roomList, room.ToSerializable()) // ✅ 변환된 데이터 추가
	}

	// ✅ 로그 추가: 반환할 JSON 확인
	log.Printf("🚀 [API Response] 방 목록 반환: %+v", roomList)

	if len(roomList) == 0 {
		// ✅ 빈 배열 반환 방지: 최소한 빈 리스트 반환
		log.Println("🚨 [Warning] 방 목록이 비어 있음. 빈 배열 반환")
		c.JSON(http.StatusOK, gin.H{"rooms": []models.SerializableRoom{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": roomList}) // ✅ JSON 응답 반환
}

// ✅ 방 삭제 핸들러 (`DELETE /delete-room/:room_id`)
func DeleteRoomHandler(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 방 ID"})
		return
	}

	room, exists := rooms[roomID]
	if !exists || room == nil { // ✅ 방이 존재하는지 확인
		c.JSON(http.StatusNotFound, gin.H{"error": "해당 방이 존재하지 않습니다."})
		return
	}

	// ✅ 방 삭제 전에 모든 연결 닫기
	room.CloseAllConnections()
	delete(rooms, roomID)

	log.Printf("채팅방 삭제 성공: %d", roomID)
	c.JSON(http.StatusOK, gin.H{"message": "채팅방 삭제 성공"})
}

// ✅ 고유 ID 생성 함수
func generateUniqueID() string {
	return time.Now().Format("20060102150405.000")
}
