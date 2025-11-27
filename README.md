# chat-app
A real-time chat application with a Go backend and Angular frontend, supporting multiple rooms and live messaging via WebSockets.

# ############Features###########
->Real-time messaging using WebSockets
->Multiple chat rooms
->Username-based identification
->Join/leave notifications
->Health check endpoint
->CORS enabled for frontend-backend integration

# ###########Backend (Go)##########
# Requirements
-> Go 1.21+
-> Gorilla WebSocket package (github.com/gorilla/websocket)
# Run the Backend
    cd backend
    go mod tidy
    go run main.go

Server runs on http://localhost:8080

# Endpoints
| Endpoint  | Method    | Description                                                  |
| --------- | --------- | ------------------------------------------------------------ |
| `/ws`     | WebSocket | Connect to the chat server. Query params: `username`, `room` |
| `/health` | GET       | Health check endpoint                                        |

# ##########Websocket Example##########
const socket = new WebSocket("ws://localhost:8080/ws?username=Alice&room=general");

socket.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(msg);
};

socket.onopen = () => {
  socket.send(JSON.stringify({ type: "message", content: "Hello everyone!" }));
};

# (You can run this example on Developer tools)

# ######################### Frontend (Angular)#####################################
Run the frontend
cd frontend
npm install
ng serve -o
Open http://localhost:4200

# ##########Project Structure##############
chat-app/
├─ backend/       # Go WebSocket server
│  └─ main.go
├─ frontend/      # Angular frontend
│  └─ src/app/chat/  # Chat component
├─ README.md
├─ .gitignore
└─ package.json


# ##############Usage###########
Start the Go backend
Start the Angular frontend
Open multiple browser tabs to simulate multiple users
Send messages and see them appear in real-time
