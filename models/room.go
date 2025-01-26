package models

type Participant struct {
	UserID   string
	UserName string
}

type Room struct {
	RoomID       int
	RoomName     string
	CreaterID    string
	Participants []Participant
}
