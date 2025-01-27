// package main

// import (
// 	"Go-chat/db"     // db 패키지
// 	"Go-chat/models" // models 패키지
// 	"log"
// 	"os"

// 	"github.com/joho/godotenv"
// )

// func main() {

// 	if err := godotenv.Load(); err != nil {
// 		log.Fatalf(".env 파일 로드 실패: %v", err)
// 	}

// 	// 환경 변수에서 MongoDB URI 가져오기
// 	mongoURI := os.Getenv("MONGO_URI")

// 	// MongoDB 초기화
// 	if err := db.InitMongoDB(mongoURI); err != nil {
// 		log.Fatalf("MongoDB 초기화 실패: %v", err)
// 	}
// 	defer func() {
// 		if err := db.MongoClient.Disconnect(nil); err != nil {
// 			log.Printf("MongoDB 연결 종료 실패: %v", err)
// 		}
// 	}()

// 	user := models.User{
// 		UserID:   "K12kd",
// 		UserName: "카카로트",
// 	}

// 	room := models.Room{
// 		RoomID:    1,
// 		RoomName:  "고를 공부하나요",
// 		CreaterID: "카카로트",
// 		Participants: []models.Participant{
// 			{
// 				UserID:   "1",
// 				UserName: "푸푸",
// 			},
// 		},
// 	}

// 	message := models.Message{
// 		MessageID: "lK123",
// 		RoomID:    1,
// 		UserID:    "K12kd",
// 		Content:   "만나서 반갑습니다 여러분!",
// 		Timestamp: "2025-01-26",
// 	}
// 	log.Printf("User: %+v\n", user)
// 	log.Printf("Room: %+v\n", room)
// 	log.Printf("Message: %+v\n", message)
// 	// User 데이터 저장
// 	if err := db.InsertDocument("GoChatDB", "users", user); err != nil {
// 		log.Printf("User 저장 실패: %v", err)
// 	}
// 	// Room 데이터 저장
// 	if err := db.InsertDocument("GoChatDB", "rooms", room); err != nil {
// 		log.Printf("Room 저장 실패: %v", err)
// 	}
// 	if err := db.InsertDocument("GoChatDB", "messages", message); err != nil {
// 		log.Printf("Message 저장 실패: %v", err)
// 	}
// }

package main

import (
	"Go-chat/websocket"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/static", "./static")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 프론트엔드 도메인
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 방 생성 API
	r.POST("/create-room", websocket.CreateRoomHandler)

	// 방 삭제 API
	r.DELETE("/delete-room/:room_id", websocket.DeleteRoomHandler)

	// 방 목록 반환 API
	r.GET("/list-rooms", websocket.ListRoomsHandler)

	// WebSocket 핸들러
	r.GET("/ws", websocket.WebSocketHandler)

	// 서버 실행
	r.Run(":8080") // http://localhost:8080
}
