package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
)

func TestRideRequestRespond_NegotiateCreatesChatConversation(t *testing.T) {
	st := newTestStore(t)

	rider, err := st.CreateOAuthUser(context.Background(), "Rider", "chat-rider@example.com", "google")
	if err != nil {
		t.Fatalf("create rider: %v", err)
	}
	driver, err := st.CreateOAuthUser(context.Background(), "Driver", "chat-driver@example.com", "google")
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}

	req, err := st.CreateRideRequest(context.Background(), models.RideRequest{
		UserID:            rider.ID,
		FromLabel:         "Atlanta",
		FromLat:           33.7490,
		FromLon:           -84.3880,
		ToLabel:           "Savannah",
		ToLat:             32.0809,
		ToLon:             -81.0912,
		RideDate:          "2026-05-20",
		RideTime:          "10:00",
		Flexibility:       "exact",
		Passengers:        1,
		Luggage:           "none",
		MaxBudget:         40,
		RideType:          "shared",
		VehiclePreference: "any",
	})
	if err != nil {
		t.Fatalf("create ride request: %v", err)
	}

	h := &AuthHandler{store: st}
	payload := map[string]any{
		"action":  "negotiate",
		"message": "Can we shift pickup by 15 minutes?",
	}
	body, _ := json.Marshal(payload)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/ride-requests/"+req.ID+"/respond", bytes.NewReader(body))
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), middleware.UserIDContextKey, driver.ID))
	rec := httptest.NewRecorder()

	h.RideRequestRespond(rec, httpReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	conversations, err := st.ListChatSummariesByUser(context.Background(), rider.ID)
	if err != nil {
		t.Fatalf("list conversations: %v", err)
	}
	if len(conversations) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(conversations))
	}
	if conversations[0].RideRequestID != req.ID {
		t.Fatalf("expected rideRequestId=%s got %s", req.ID, conversations[0].RideRequestID)
	}

	messages, err := st.ListChatMessagesByConversation(context.Background(), conversations[0].ID, rider.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Body != "Can we shift pickup by 15 minutes?" {
		t.Fatalf("unexpected first message: %s", messages[0].Body)
	}
}

func TestChats_SendAndListMessages(t *testing.T) {
	st := newTestStore(t)

	rider, err := st.CreateOAuthUser(context.Background(), "Rider", "chat2-rider@example.com", "google")
	if err != nil {
		t.Fatalf("create rider: %v", err)
	}
	driver, err := st.CreateOAuthUser(context.Background(), "Driver", "chat2-driver@example.com", "google")
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}

	req, err := st.CreateRideRequest(context.Background(), models.RideRequest{
		UserID:            rider.ID,
		FromLabel:         "Charlotte",
		FromLat:           35.2271,
		FromLon:           -80.8431,
		ToLabel:           "Raleigh",
		ToLat:             35.7796,
		ToLon:             -78.6382,
		RideDate:          "2026-06-12",
		RideTime:          "09:00",
		Flexibility:       "exact",
		Passengers:        1,
		Luggage:           "small",
		MaxBudget:         30,
		RideType:          "shared",
		VehiclePreference: "any",
	})
	if err != nil {
		t.Fatalf("create ride request: %v", err)
	}

	conversation, err := st.CreateOrGetConversation(context.Background(), rider.ID, driver.ID, req.ID, "", "")
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	h := &AuthHandler{store: st}

	sendBody := bytes.NewBufferString(`{"body":"Hello from the rider"}`)
	sendReq := httptest.NewRequest(http.MethodPost, "/api/chats/"+conversation.ID+"/messages", sendBody)
	sendReq = sendReq.WithContext(context.WithValue(sendReq.Context(), middleware.UserIDContextKey, rider.ID))
	sendRec := httptest.NewRecorder()

	h.ChatByID(sendRec, sendReq)

	if sendRec.Code != http.StatusCreated {
		t.Fatalf("send status=%d body=%s", sendRec.Code, sendRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/chats/"+conversation.ID+"/messages", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), middleware.UserIDContextKey, driver.ID))
	listRec := httptest.NewRecorder()

	h.ChatByID(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRec.Code, listRec.Body.String())
	}

	var res struct {
		Messages []models.ChatMessage `json:"messages"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal messages: %v", err)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(res.Messages))
	}
	if res.Messages[0].Body != "Hello from the rider" {
		t.Fatalf("unexpected message body: %s", res.Messages[0].Body)
	}

	driverNotes, err := st.ListNotificationsByUser(context.Background(), driver.ID)
	if err != nil {
		t.Fatalf("list driver notifications: %v", err)
	}
	if len(driverNotes) == 0 {
		t.Fatalf("expected notification for driver after chat message")
	}
}

func TestChats_SendImageAndLocationMessages(t *testing.T) {
	st := newTestStore(t)

	rider, err := st.CreateOAuthUser(context.Background(), "Rider", "chat3-rider@example.com", "google")
	if err != nil {
		t.Fatalf("create rider: %v", err)
	}
	driver, err := st.CreateOAuthUser(context.Background(), "Driver", "chat3-driver@example.com", "google")
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}

	req, err := st.CreateRideRequest(context.Background(), models.RideRequest{
		UserID:            rider.ID,
		FromLabel:         "Miami",
		FromLat:           25.7617,
		FromLon:           -80.1918,
		ToLabel:           "Orlando",
		ToLat:             28.5383,
		ToLon:             -81.3792,
		RideDate:          "2026-06-12",
		RideTime:          "09:00",
		Flexibility:       "exact",
		Passengers:        1,
		Luggage:           "small",
		MaxBudget:         30,
		RideType:          "shared",
		VehiclePreference: "any",
	})
	if err != nil {
		t.Fatalf("create ride request: %v", err)
	}

	conversation, err := st.CreateOrGetConversation(context.Background(), rider.ID, driver.ID, req.ID, "", "")
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	h := &AuthHandler{store: st}

	imageReq := httptest.NewRequest(http.MethodPost, "/api/chats/"+conversation.ID+"/messages", bytes.NewBufferString(`{"messageType":"image","body":"Pickup spot photo","imageData":"data:image/png;base64,abc123"}`))
	imageReq = imageReq.WithContext(context.WithValue(imageReq.Context(), middleware.UserIDContextKey, rider.ID))
	imageRec := httptest.NewRecorder()
	h.ChatByID(imageRec, imageReq)
	if imageRec.Code != http.StatusCreated {
		t.Fatalf("image status=%d body=%s", imageRec.Code, imageRec.Body.String())
	}

	locationReq := httptest.NewRequest(http.MethodPost, "/api/chats/"+conversation.ID+"/messages", bytes.NewBufferString(`{"messageType":"location","body":"Meet here","locationLabel":"Airport Terminal A","locationLat":28.4312,"locationLon":-81.3081}`))
	locationReq = locationReq.WithContext(context.WithValue(locationReq.Context(), middleware.UserIDContextKey, driver.ID))
	locationRec := httptest.NewRecorder()
	h.ChatByID(locationRec, locationReq)
	if locationRec.Code != http.StatusCreated {
		t.Fatalf("location status=%d body=%s", locationRec.Code, locationRec.Body.String())
	}

	messages, err := st.ListChatMessagesByConversation(context.Background(), conversation.ID, rider.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].MessageType != "image" || messages[0].ImageData == "" {
		t.Fatalf("expected first message to be image, got %+v", messages[0])
	}
	if messages[1].MessageType != "location" || messages[1].LocationLabel != "Airport Terminal A" {
		t.Fatalf("expected second message to be location, got %+v", messages[1])
	}
}
