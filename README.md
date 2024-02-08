**This is a realtime chat server that allows 2 major operation:**
1. Private Chat between 2 clients
2. Broadcast chat between multiple clients

**Prerequisties to Run the program:**
1. Should have installed GO on your system.
2. Use any editor like VS Code or anything or use the terminal/cmd to execute the program.
3. Should have installed any application to execute the Websocket APIs like Postman.


**How to Run the program:**
1. Clone the repository on your local system.
2. Now open the terminal/cmd and Go to the path where main.go file is present.
3. Run the command
   > **go run main.go**

The above command will run the server on Port **8888**.
If you have already used this Port then kindly go to the **main.go** program and change the port on below line
   > **server := &http.Server{
		Addr:    ":8888",
		Handler: routes,
	}**



**Test the scenario --> Private Chat between 2 clients**
1. Open the postman --> go to New --> choose WebSocket
2. Now to chat privately use the below URL and open two tab: **(Please verify the port before using the url if you have changed the port on previous steps)**

	**Tab1---> ws://localhost:8888/wschat/private?sender=user1** (click on Connect and verify the connection is succesfull or not)

	message for first user--> Here on **receiver** you have to provide the **user2** name.

	> **{
    	"receiver": "user2",
    	"text": "Hi, how are you!"
	}**

	**Tab2---> ws://localhost:8888/wschat/private?sender=user2** (click on Connect and verify the connection is succesfull or not)

	message for second user--> Here on **receiver** you have to provide the **user1** name.

	> **{
    	"receiver": "user1",
    	"text": "Hi, I am good!"
	}**

4. After connecting and providing the message as mentioned on previous steps now click on send.
5. Now the users will be able to see the text message send by each other on response.



**Test the scenario --> Broadcast chat between multiple clients**
1. Open the postman --> go to New --> choose WebSocket
2. Now to boradcast chat use the below URL and open three tab: **(Please verify the port before using the url if you have changed the port already)**
   
	**Tab1---> ws://localhost:8888/wschat/broadcast?sender=user1** (click on Connect and verify the connection is succesfull or not)
	
 	message for first user--> Here you don't have to worry about the receiver as it will boradcasting to all.
	
 	> **{
   	"receiver": "user2",
    	"text": "Hi, how are you!"
	}**

	**Tab2---> ws://localhost:8888/wschat/broadcast?sender=user2** (click on Connect and verify the connection is succesfull or not)

	**Tab2---> ws://localhost:8888/wschat/broadcast?sender=user3** (click on Connect and verify the connection is succesfull or not)

4. After connecting all the users and providing the message for first user as mentioned on previous steps now click on send for first user on Tab1.
5. Now the all three users will be able to see the text message send by each other on response.


**How to run the unit test cases:**
1. open the terminal/cmd and Go to the path where **main_test.go** file is present.
2. Run the command to execute first test.
   > **go test -run TestPrivateChat -v**
3. Run the command to execute second test.
   > **go test -run TestBroadcastChat -v**
   

   
