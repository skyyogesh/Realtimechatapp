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
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
}

func TestPrivateChat(t *testing.T) {
	router := gin.Default()

	router.GET("/wschat", controller.PrivateChat)

	server := httptest.NewServer(router)
	defer server.Close()

	conn1, err := connectClient("user1", server, "private")
	assert.NoError(t, err)
	defer conn1.Close()

	conn2, err := connectClient("user2", server, "private")
	assert.NoError(t, err)
	defer conn2.Close()

	conn3, err := connectClient("user3", server, "private")
	assert.NoError(t, err)
	defer conn3.Close()

	// create a JSON-encoded message for conn1 to send it to conn2
	message1 := Message{Sender: "user1", Receiver: "user2", Text: "This is private chat test 1"}
	// create a JSON-encoded message for conn1 to send it to conn3
	message2 := Message{Sender: "user1", Receiver: "user3", Text: "This is private chat test 2"}

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

	router.GET("/wschat/broadcast", controller.Broadcast)

	server := httptest.NewServer(router)
	defer server.Close()

	conn1, err := connectClient("broaduser1", server, "broadcast")
	assert.NoError(t, err)
	defer conn1.Close()

	conn2, err := connectClient("broaduser2", server, "broadcast")
	assert.NoError(t, err)
	defer conn2.Close()

	conn3, err := connectClient("broaduser3", server, "broadcast")
	assert.NoError(t, err)
	defer conn3.Close()

	// Send a JSON-encoded message from conn1 and assert that conn2 & conn3 receives it
	message := Message{Sender: "broaduser1", Receiver: "broaduser2", Text: "This is broadcast chat test"}
	jsonMessage, err := json.Marshal(message)
	assert.NoError(t, err)

	// Send a message from conn1 and assert that conn2 receives it
	sendMessage(conn1, jsonMessage)
	conn2receivedMessage := receiveMessage(conn2)
	conn3receivedMessage := receiveMessage(conn3)
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

func connectClient(username string, server *httptest.Server, chattype string) (*websocket.Conn, error) {
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	if chattype == "private" {
		u.Path = "/wschat"
		query := url.Values{"username": {username}}
		u.RawQuery = query.Encode()
	} else {
		u.Path = "/wschat/broadcast"
	}

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	return conn, err
}
