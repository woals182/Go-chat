package websocket

import (
	"Go-chat/models"
	"fmt"
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

var rooms = make(map[int]*Room) // Room ID -> Room 매핑

// WebSocket 핸들러 (방 선택 가능)
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("웹소켓 업그레이드 실패: %v", err)
		return
	}
	defer conn.Close()

	roomIDStr := c.Query("room_id")
	if roomIDStr == "" {
		log.Printf("방 ID가 제공되지 않았습니다.")
		c.JSON(http.StatusBadRequest, gin.H{"error": "방 ID가 필요합니다."})
		return
	}
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

// 방 생성 핸들러
func CreateRoomHandler(c *gin.Context) {
	var request struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "방 이름을 제공해야 합니다."})
		return
	}

	roomID := len(rooms) + 1
	if _, exists := rooms[roomID]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "이미 존재하는 방입니다."})
		return
	}

	rooms[roomID] = NewRoom(roomID, request.Name)
	log.Printf("새 채팅방 생성: %s (ID: %d)", request.Name, roomID)

	c.JSON(http.StatusOK, gin.H{
		"room_id":   roomID,
		"room_name": request.Name,
	})
}

// 방 목록 반환 핸들러
func ListRoomsHandler(c *gin.Context) {
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

	c.JSON(http.StatusOK, roomList)
}

// 방 삭제 핸들러
func DeleteRoomHandler(c *gin.Context) {
	log.Println("DeleteRoomHandler 호출됨")

	roomIDStr := c.Param("room_id")
	if roomIDStr == "" {
		log.Println("방 ID를 제공하지 않음")
		c.JSON(http.StatusBadRequest, gin.H{"error": "방 ID를 제공해야 합니다."})
		return
	}

	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		log.Printf("유효하지 않은 방 ID: %s", roomIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "유효하지 않은 방 ID입니다."})
		return
	}

	room, exists := rooms[roomID]
	if !exists {
		log.Printf("방 ID %d가 존재하지 않음", roomID)
		c.JSON(http.StatusNotFound, gin.H{"error": "해당 방이 존재하지 않습니다."})
		return
	}

	room.CloseAllConnections()
	delete(rooms, roomID)

	log.Printf("채팅방 삭제 성공: %d", roomID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("채팅방 %d 삭제 성공", roomID)})
}

// 고유 ID 생성
func generateUniqueID() string {
	return time.Now().Format("20060102150405.000")
}
