# Battleship-Go 🚢

A real-time, multiplayer Battleship game built with **Go** (Golang) and **WebSockets**. Experience the classic naval combat strategy game directly in your browser with seamless matchmaking and live gameplay.

![Battleship-Go Banner](https://via.placeholder.com/800x200?text=Battleship-Go+Multiplayer)

## 🌟 Features

-   **Real-Time Multiplayer**: Challenge other players instantly via the live lobby.
-   **WebSocket Communication**: Fast, low-latency updates for game state, shots, and chat.
-   **Interactive UI**:
    -   **Lobby**: See who's online and send challenge requests.
    -   **Ship Placement**: Drag-and-drop or click-to-place interface with rotation support.
    -   **Battle Interface**: Clear view of your board and the enemy's waters with hit/miss indicators.
-   **Game Logic**: Full server-side validation for turns, ship placement, and hit detection.
-   **Single Page Application (SPA)**: Smooth transitions between lobby, placement, and game modes without reloading.

## 🛠️ Tech Stack

-   **Backend**: Go (Golang)
    -   `gorilla/websocket` for real-time communication.
    -   `google/uuid` for unique IDs.
-   **Frontend**: HTML5, CSS3, Vanilla JavaScript.
    -   No heavy frameworks, just pure, fast web technologies.

## 🚀 Getting Started

### Prerequisites

-   [Go](https://go.dev/dl/) (version 1.18 or higher recommended)
-   A modern web browser.

### Installation

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/Souptim/battleship-go.git
    cd battleship-go
    ```

2.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

### Running the Game

1.  **Start the server**:
    ```bash
    go run cmd/server/main.go
    ```

2.  **Open the game**:
    -   Open your browser and navigate to `http://localhost:8080`.
    -   Open a second tab (or use a different device on the same network) to simulate a second player.

## 🎮 How to Play

1.  **Join the Lobby**: Enter your name to connect.
2.  **Challenge a Player**: Click "Challenge" next to another player's name in the lobby list.
3.  **Place Ships**:
    -   You have 5 ships: Carrier (5), Battleship (4), Cruiser (3), Submarine (3), Destroyer (2).
    -   Click on the grid to place them.
    -   Use the **"H" / "V"** button to toggle orientation.
    -   Click **"Send Ships"** when ready.
4.  **Battle**:
    -   Take turns firing at the enemy grid.
    -   **Red** marker = HIT.
    -   **White/Grey** marker = MISS.
    -   Sink all 5 enemy ships to win!

## 📂 Project Structure

```
battleship-go/
├── cmd/
│   └── server/         # Entry point for the application
├── internal/
│   ├── game/           # Core game logic and types
│   └── ws/             # WebSocket handlers and connection management
├── web/                # Frontend assets (HTML, CSS, JS)
├── go.mod              # Go module definition
└── README.md           # Project documentation
```

## 🤝 Contributing

Contributions are welcome! Feel free to submit a Pull Request.

1.  Fork the project.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## 📄 License

This project is open source and available under the [MIT License](LICENSE).
