package models

import "time"

type ChatConversation struct {
	ID              string    `json:"id"`
	RiderUserID     string    `json:"riderUserId"`
	DriverUserID    string    `json:"driverUserId"`
	RideRequestID   string    `json:"rideRequestId"`
	PublishedRideID string    `json:"publishedRideId"`
	TripID          string    `json:"tripId"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type ChatMessage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	SenderUserID   string    `json:"senderUserId"`
	Body           string    `json:"body"`
	MessageType    string    `json:"messageType"`
	ImageData      string    `json:"imageData"`
	LocationLabel  string    `json:"locationLabel"`
	LocationLat    float64   `json:"locationLat"`
	LocationLon    float64   `json:"locationLon"`
	CreatedAt      time.Time `json:"createdAt"`
}

type ChatConversationSummary struct {
	ID                  string    `json:"id"`
	RiderUserID         string    `json:"riderUserId"`
	DriverUserID        string    `json:"driverUserId"`
	RideRequestID       string    `json:"rideRequestId"`
	PublishedRideID     string    `json:"publishedRideId"`
	TripID              string    `json:"tripId"`
	Status              string    `json:"status"`
	TripStatus          string    `json:"tripStatus"`
	OtherUserID         string    `json:"otherUserId"`
	OtherUserName       string    `json:"otherUserName"`
	OtherUserAvatarData string    `json:"otherUserAvatarData"`
	FromLabel           string    `json:"fromLabel"`
	ToLabel             string    `json:"toLabel"`
	RideDate            string    `json:"rideDate"`
	RideTime            string    `json:"rideTime"`
	LastMessage         string    `json:"lastMessage"`
	LastMessageSenderID string    `json:"lastMessageSenderId"`
	LastMessageAt       time.Time `json:"lastMessageAt"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
