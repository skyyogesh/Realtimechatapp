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
	conn        *websocket.Conn
	sender      string
	sendMessage chan Message
}

type Message struct {
	Receiver string `json:"receiver"`
	Text     string `json:"text"`
}

var clientsMutex sync.Mutex
var wg sync.WaitGroup
var privateclients = make(map[string]*Client)
var broadcastclients = make(map[*Client]bool)
var broadcastMessage = make(chan Message)
var privateMessage = make(chan Message)

const private string = "private"
const broadcast string = "broadcast"

func PrivateChat(c *gin.Context) {
	MessageHandler(c, private)
}

func Broadcast(c *gin.Context) {
	MessageHandler(c, broadcast)
}

func MessageHandler(c *gin.Context, chattype string) {
	client := handleWsChatConnection(c)
	if client == nil {
		fmt.Println("Invalid URL:", c.Request.URL)
		return
	}

	if chattype == private {
		sender := c.Query("sender")
		if sender == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sender is required"})
			return
		}
		clientsMutex.Lock()
		privateclients[sender] = client
		clientsMutex.Unlock()
	} else {
		clientsMutex.Lock()
		broadcastclients[client] = true
		clientsMutex.Unlock()
	}

	wg.Add(3)
	go client.readMessages(chattype, &wg)
	go client.writeMessages(chattype, &wg)
	go handleMessages(chattype, &wg)
	wg.Wait()
}

func handleWsChatConnection(c *gin.Context) (client *Client) {
	sender := c.Query("sender")
	if sender == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade connection to WebSocket"})
		return
	}

	client = &Client{conn: conn, sender: sender, sendMessage: make(chan Message)}
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

		if chattype == private {
			privateMessage <- message
		} else {
			broadcastMessage <- message
		}
	}
}

func (client *Client) writeMessages(chattype string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer client.closeConnection(chattype)

	for {
		chatmessage, ok := <-client.sendMessage
		if !ok {
			fmt.Println("Channel is closed.")
			break
		}

		err := client.conn.WriteMessage(websocket.TextMessage, []byte(chatmessage.Text))
		if err != nil {
			fmt.Println("Error while writing message to the client")
			break
		}
	}
}

func (client *Client) closeConnection(chattype string) {
	client.conn.Close()
	if chattype == private {
		clientsMutex.Lock()
		delete(privateclients, client.sender)
		clientsMutex.Unlock()
	} else {
		clientsMutex.Lock()
		delete(broadcastclients, client)
		clientsMutex.Unlock()
	}
}

func handleMessages(chattype string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		if chattype == private {
			chatmessage := <-privateMessage
			receiver, exists := privateclients[chatmessage.Receiver]
			if exists {
				select {
				case receiver.sendMessage <- chatmessage:
				default:
					close(receiver.sendMessage)
					clientsMutex.Lock()
					delete(privateclients, receiver.sender)
					clientsMutex.Unlock()
				}
			} else {
				var userlist []string
				for user := range privateclients {
					clientsMutex.Lock()
					userlist = append(userlist, user)
					clientsMutex.Unlock()
				}
				fmt.Printf("Receiver %v is currently Offilne !!\n", chatmessage.Receiver)
				fmt.Printf("Please connect with available Online user %v\n", userlist)
			}
		} else {
			chatmessage := <-broadcastMessage
			for client := range broadcastclients {
				select {
				case client.sendMessage <- chatmessage:
				default:
					close(client.sendMessage)
					clientsMutex.Lock()
					delete(broadcastclients, client)
					clientsMutex.Unlock()
				}
			}

		}
	}
}
