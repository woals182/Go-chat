package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "sync"
	"time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

// Message 구조체는 클라이언트 간 메시지 전달을 담당합니다.
type Message struct {
    RoomId    int `json:"roomId"`
    RoomName  string `json:"roomName"`
    OwnerName string `json:"ownerName"`
    Username  string `json:"username"`
    Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

// 싱글톤 패턴: 애플리케이션 전역에서 공유하는 데이터 구조
var rooms = make(map[string]map[*websocket.Conn]string) // 각 방의 클라이언트 연결
var roomMessages = make(map[string][]Message)          // 각 방의 메시지 기록
var broadcast = make(chan Message)                    // 퍼블리셔-서브스크라이버 패턴으로 메시지를 전달
var mutex = &sync.Mutex{}                             // 뮤텍스 패턴으로 동시성 제어

func main() {
    r := gin.Default()

    // 방 생성 엔드포인트 등록
    r.POST("/create-room", func(c *gin.Context) {
        var req struct {
            RoomId    int    `json:"roomId"`
            RoomName  string `json:"roomName"`
            OwnerName string `json:"ownerName"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
            return
        }

        mutex.Lock()
        defer mutex.Unlock()

        if _, exists := rooms[req.RoomName]; exists {
            c.JSON(http.StatusConflict, gin.H{"error": "Room already exists"})
            return
        }

        rooms[req.RoomName] = make(map[*websocket.Conn]string) // 방 생성
		fmt.Print(rooms)

        roomMessages[req.RoomName] = []Message{}              // 메시지 초기화

        fmt.Printf("채팅방 '%s' (ID: %d) 생성됨 (주인: %s)\n", req.RoomName, req.RoomId, req.OwnerName)
        c.JSON(http.StatusOK, gin.H{"message": "Room created successfully"})
    })

    // 방 목록 가져오기 엔드포인트 등록
    r.GET("/rooms", func(c *gin.Context) {
        mutex.Lock()
        defer mutex.Unlock()

        // 방 목록 생성
        roomList := []map[string]string{}
        for roomName := range rooms {
            roomList = append(roomList, map[string]string{
                "roomName": roomName,
            })
        }

        c.JSON(http.StatusOK, roomList)
    })

    // WebSocket 엔드포인트 등록
    r.GET("/ws", func(c *gin.Context) {
        handleConnections(c.Writer, c.Request)
    })

    // 메시지 브로드캐스트를 처리하는 고루틴 실행
    go handleMessages()

    r.Static("/static", "./static")
    r.Run(":8080")
}


func formatTimeForChat(t time.Time) string {
    // 시간 포맷 변환
    hour := t.Hour()
    period := "오전"
    if hour >= 12 {
        period = "오후"
        if hour > 12 {
            hour -= 12
        }
    }
    return fmt.Sprintf("%s %d:%02d", period, hour, t.Minute())
}

// 클라이언트 연결을 처리하는 함수
func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("WebSocket 업그레이드 실패:", err)
        return
    }
    defer ws.Close()

    var initMsg Message
    // 클라이언트 초기 메시지 수신
    err = ws.ReadJSON(&initMsg)
    if err != nil {
        fmt.Println("초기 메시지 수신 실패:", err)
        return
    }
	roomId := initMsg.RoomId
    roomName := initMsg.RoomName
    username := initMsg.Username

    // 방 생성 및 사용자 등록 (싱글톤 패턴 + 뮤텍스 패턴)
    mutex.Lock()
    if _, ok := rooms[roomName]; !ok {
        rooms[roomName] = make(map[*websocket.Conn]string)
        roomMessages[roomName] = []Message{}
        fmt.Printf("채팅방 '%s' 생성됨 (주인: %s)\n", roomName, initMsg.OwnerName)
    }
    rooms[roomName][ws] = username
    mutex.Unlock()

    fmt.Printf("사용자 '%s'가 채팅방 '%s'에 접속\n", username, roomName)

    // 퇴장 처리
    defer func() {
        mutex.Lock()
        delete(rooms[roomName], ws)
        mutex.Unlock()

        exitMsg := Message{
			RoomId: roomId,
            RoomName:  roomName,
            OwnerName: initMsg.OwnerName,
            Username:  "System",
            Content:   fmt.Sprintf("%s님이 퇴장하셨습니다.", username),
        }
        broadcast <- exitMsg

        if err := saveRoomMessagesToFile(roomId, roomName); err != nil {
            fmt.Printf("퇴장 시 JSON 저장 실패: %s\n", err)
        } else {
            fmt.Printf("퇴장 시 채팅방 '%s'의 메시지가 저장되었습니다.\n", roomName)
        }
    }()

    // 클라이언트 메시지 처리 루프
    for {
        var msg Message
        err := ws.ReadJSON(&msg)
        if err != nil {
            if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
                fmt.Println("클라이언트가 정상적으로 연결을 종료했습니다.")
            } else {
                fmt.Printf("메시지 읽기 오류: %s\n", err)
            }
            break
        }

        msg.RoomId = roomId // 각 메시지에 roomId 추가
        msg.RoomName = roomName
        msg.Username = username
        msg.Timestamp = formatTimeForChat(time.Now())

        mutex.Lock()
        roomMessages[roomName] = append(roomMessages[roomName], msg)
        mutex.Unlock()

        broadcast <- msg
    }
}

// 브로드캐스트 처리 (퍼블리셔-서브스크라이버 패턴)
func handleMessages() {
    for {
        msg := <-broadcast
        roomName := msg.RoomName

        mutex.Lock()
        for client := range rooms[roomName] {
            err := client.WriteJSON(msg)
            if err != nil {
                fmt.Printf("메시지 전송 오류: %s\n", err)
                client.Close()
                delete(rooms[roomName], client)
            }
        }
        mutex.Unlock()
    }
}

// JSON 파일에 메시지 저장 (템플릿 메서드 패턴과 유사한 구조 적용 가능)
func saveRoomMessagesToFile(roomId int, roomName string) error {
    mutex.Lock()
    defer mutex.Unlock()

    messages, ok := roomMessages[roomName]
    if !ok || len(messages) == 0 {
        return fmt.Errorf("채팅방 '%s'의 메시지가 비어 있습니다", roomName)
    }

    fileName := fmt.Sprintf("%s_messages.json", roomId, roomName)
    file, err := os.Create(fileName)
    if err != nil {
        return fmt.Errorf("파일 생성 실패: %w", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(messages); err != nil {
        return fmt.Errorf("JSON 인코딩 실패: %w", err)
    }

    fmt.Printf("채팅방 '%s'의 메시지가 %s 파일에 저장되었습니다.\n", roomName, roomId, fileName)
    return nil
}
