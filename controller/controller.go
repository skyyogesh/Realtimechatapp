package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	username string
	send     chan []byte
}

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
}

var clientsMutex sync.Mutex
var privateclients = make(map[string]*Client)
var broadcastclients = make(map[*Client]bool)
var broadcast = make(chan Message)

func PrivateChat(c *gin.Context) {
	MessageHandler(c, "private")
}

func Broadcast(c *gin.Context) {
	MessageHandler(c, "broadcast")
}

func MessageHandler(c *gin.Context, chattype string) {
	var wg sync.WaitGroup
	client := handleWsChatConnection(c)

	if chattype == "private" {
		// Get username from query parameter
		client.username = c.Query("username")
		if client.username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
			return
		}
		clientsMutex.Lock()
		privateclients[client.username] = client
		clientsMutex.Unlock()
	} else {
		clientsMutex.Lock()
		broadcastclients[client] = true
		clientsMutex.Unlock()
	}

	wg.Add(2)
	go client.readMessages(chattype, &wg)
	go client.writeMessages(chattype, &wg)
	go HandleMessages(chattype, &wg)
	wg.Wait()
}

func handleWsChatConnection(c *gin.Context) (client *Client) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection to WebSocket"})
		return
	}

	client = &Client{conn: conn, send: make(chan []byte)}
	return client
}

func (client *Client) readMessages(chattype string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer client.closeConnection(chattype)

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		message := Message{}

		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Println("Error :", err)
			return
		}

		broadcast <- message
	}
}

func (client *Client) writeMessages(chattype string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer client.closeConnection(chattype)

	for {
		msg, ok := <-client.send
		if !ok {
			fmt.Println("Channel is closed.")
			break
		}

		chatmessage := Message{}
		if err := json.Unmarshal(msg, &chatmessage); err != nil {
			fmt.Println("Error :", err)
			return
		}
		// Broadcasting only the text message
		err := client.conn.WriteMessage(websocket.TextMessage, []byte(chatmessage.Text))
		if err != nil {
			fmt.Println("Error while writing message to the client")
			break
		}
	}
}

func (client *Client) closeConnection(chattype string) {
	client.conn.Close()
	if chattype == "private" {
		clientsMutex.Lock()
		delete(privateclients, client.username)
		clientsMutex.Unlock()
	} else {
		clientsMutex.Lock()
		delete(broadcastclients, client)
		clientsMutex.Unlock()
	}
}

func HandleMessages(chattype string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		message := <-broadcast
		recieved_message, err := json.Marshal(message)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}

		if chattype == "private" {
			receiver, exists := privateclients[message.Receiver]
			if exists {
				select {
				case receiver.send <- recieved_message:
				default:
					close(receiver.send)
					clientsMutex.Lock()
					delete(privateclients, receiver.username)
					clientsMutex.Unlock()
				}
			} else {
				var userlist []string
				for user := range privateclients {
					clientsMutex.Lock()
					userlist = append(userlist, user)
					clientsMutex.Unlock()
				}
				fmt.Printf("Receiver %v is currently Offilne !!\n", message.Receiver)
				fmt.Printf("Please connect with available Online user %v\n", userlist)
			}
		} else {
			for client := range broadcastclients {
				select {
				case client.send <- recieved_message:
				default:
					close(client.send)
					clientsMutex.Lock()
					delete(broadcastclients, client)
					clientsMutex.Unlock()
				}
			}

		}
	}
}
