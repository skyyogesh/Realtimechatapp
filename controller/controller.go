package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	Receiver string `json:"receiver" validate:"omitempty"`
	Text     string `json:"text" validate:"required"`
}

type CustomError struct {
	errMessage string
}

var clientsMutex sync.Mutex
var wg sync.WaitGroup
var broadcastclients = make(map[*Client]bool)
var broadcastMessage = make(chan Message)
var privateMessage = make(chan Message)

func RealtimeChat(c *gin.Context) {
	client := handleWsChatConnection(c)
	if client == nil {
		fmt.Println("Invalid URL:", c.Request.URL)
		return
	}

	clientsMutex.Lock()
	broadcastclients[client] = true
	clientsMutex.Unlock()

	wg.Add(3)
	go client.readMessages(&wg)
	go client.writeMessages(&wg)
	go handleMessages(&wg)
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

func (client *Client) readMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	defer client.closeConnection()

	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		message := Message{}

		if err := json.Unmarshal(msg, &message); err != nil {
			err = writeCustomError(client, "Please provide 'Receiver' & 'Text' Field to start chat")
			if err != nil {
				break
			}
			break
		}

		validate := validator.New()
		error := validate.Struct(message)
		if error != nil {
			if validationErrors, ok := error.(validator.ValidationErrors); ok {
				for _, ve := range validationErrors {
					errmsg := fmt.Sprintf("Validation error for field '%s': %s\n", ve.Field(), ve.Tag())
					err = writeCustomError(client, errmsg)
					if err != nil {
						return
					}
				}
			} else {
				log.Fatal(error)
			}
			return
		}

		if message.Receiver != "" {
			privateMessage <- message
		} else {
			broadcastMessage <- message
		}

	}
}

func (client *Client) writeMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	defer client.closeConnection()

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

func (client *Client) closeConnection() {
	client.conn.Close()
	clientsMutex.Lock()
	delete(broadcastclients, client)
	clientsMutex.Unlock()
}

func handleMessages(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case chatmessage, ok := <-broadcastMessage:
			if !ok {
				fmt.Println("Channel is closed.")
				break
			}

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

		case chatmessage, ok := <-privateMessage:
			if !ok {
				fmt.Println("Channel is closed.")
				break
			}
			for client := range broadcastclients {
				if client.sender == chatmessage.Receiver {
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
}

func writeCustomError(client *Client, errmsg string) (err error) {
	customErr := &CustomError{errMessage: errmsg}
	err = client.conn.WriteMessage(websocket.TextMessage, []byte("Error : "+customErr.errMessage))
	if err != nil {
		fmt.Println("Error while writing error message to the client")
	}
	return err
}
