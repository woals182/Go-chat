package models

type User struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

type Participant struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

type Room struct {
	RoomID       int           `json:"room_id"`
	RoomName     string        `json:"room_name"`
	CreaterID    string        `json:"creater_id"`
	Participants []Participant `json:"participants"`
}

type Message struct {
	MessageID string `json:"message_id"`
	RoomID    int    `json:"room_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}
