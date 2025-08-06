package websockets

import (
	"log"

	"github.com/gofiber/fiber/v2"
	ws "github.com/gofiber/websocket/v2"
)

func Register(app *fiber.App) {
	app.Use("/ws", ws.New(func(c *ws.Conn) {
		userID := c.Query("userId")
		if userID == "" {
			log.Println("❌ userId is required")
			c.Close()
			return
		}

		AddClient(userID, c)
		defer RemoveClient(userID)

		log.Printf("🟢 %s connected via WebSocket", userID)

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Printf("🔌 Disconnected: %s", userID)
				break
			}
		}
	}))
}
