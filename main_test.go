package main

import (
	"encoding/json"
	"log"
	"net/http/httptest"
	"net/url"
	"realtimechat/controller"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type Message struct {
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
}

func TestPrivateChat(t *testing.T) {
	router := gin.Default()

	router.GET("/wschat", controller.RealtimeChat)

	server := httptest.NewServer(router)
	defer server.Close()

	conn1, err := connectClient("user1", server)
	assert.NoError(t, err)
	defer conn1.Close()

	conn2, err := connectClient("user2", server)
	assert.NoError(t, err)
	defer conn2.Close()

	conn3, err := connectClient("user3", server)
	assert.NoError(t, err)
	defer conn3.Close()

	// create a JSON-encoded message for conn1 to send it to conn2
	message1 := Message{Receiver: "user2", Text: "This is private chat test 1"}
	// create a JSON-encoded message for conn1 to send it to conn3
	message2 := Message{Receiver: "user3", Text: "This is private chat test 2"}

	jsonMessage1, err := json.Marshal(message1)
	assert.NoError(t, err)

	jsonMessage2, err := json.Marshal(message2)
	assert.NoError(t, err)

	// Send a message from conn1 and assert that conn2 receives it
	sendMessage(conn1, jsonMessage1)
	conn2receivedMessage := receiveMessage(conn2)
	assert.Equal(t, message1.Text, conn2receivedMessage)

	// Send a message from conn1 and assert that conn3 receives it
	sendMessage(conn1, jsonMessage2)
	conn3receivedMessage := receiveMessage(conn3)
	assert.Equal(t, message2.Text, conn3receivedMessage)

}

func TestBroadcastChat(t *testing.T) {
	router := gin.Default()

	router.GET("/wschat", controller.RealtimeChat)

	server := httptest.NewServer(router)
	defer server.Close()

	broadconn1, err := connectClient("broaduser1", server)
	assert.NoError(t, err)
	defer broadconn1.Close()

	broadconn2, err := connectClient("broaduser2", server)
	assert.NoError(t, err)
	defer broadconn2.Close()

	broadconn3, err := connectClient("broaduser3", server)
	assert.NoError(t, err)
	defer broadconn3.Close()

	// Send a JSON-encoded message from conn1 and assert that conn2 & conn3 receives it
	message := Message{Text: "This is broadcast chat test"}
	jsonMessage, err := json.Marshal(message)
	assert.NoError(t, err)

	// Send a message from conn1 and assert that conn2 receives it
	sendMessage(broadconn1, jsonMessage)
	conn2receivedMessage := receiveMessage(broadconn2)
	conn3receivedMessage := receiveMessage(broadconn3)
	assert.Equal(t, message.Text, conn2receivedMessage)
	assert.Equal(t, message.Text, conn3receivedMessage)

}

func sendMessage(conn *websocket.Conn, message []byte) {
	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Fatal(err)
	}
}

func receiveMessage(conn *websocket.Conn) string {
	_, receivedMessage, err := conn.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}
	return string(receivedMessage)
}

func connectClient(sender string, server *httptest.Server) (*websocket.Conn, error) {
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.Path = "/wschat"
	query := url.Values{"sender": {sender}}
	u.RawQuery = query.Encode()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	return conn, err
}
