package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Message struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

type PublishRequest struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

type SendMessageRequest struct {
	Topic   string `json:"topic"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type TokenRequest struct {
	User string `json:"user"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type CentrifugoConfig struct {
	URL       string
	APIKey    string
	TokenHMACSecretKey string
}

var centrifugoConfig = CentrifugoConfig{
	URL:                getEnv("CENTRIFUGO_URL", "http://localhost:8000"),
	APIKey:             getEnv("CENTRIFUGO_API_KEY", "api_key"),
	TokenHMACSecretKey: getEnv("CENTRIFUGO_TOKEN_HMAC_SECRET_KEY", "token_hmac_secret_key"),
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateToken(user string) (string, error) {
	// Create the claims
	claims := jwt.MapClaims{
		"sub": user,                                    // Subject (user identifier)
		"iat": time.Now().Unix(),                      // Issued at
		"exp": time.Now().Add(24 * time.Hour).Unix(),  // Expires in 24 hours
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(centrifugoConfig.TokenHMACSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func getToken(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.User == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is required"})
		return
	}

	token, err := generateToken(req.User)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

func publishToCentrifugo(channel string, data interface{}) error {
	publishReq := PublishRequest{
		Channel: channel,
		Data:    data,
	}

	jsonData, err := json.Marshal(publishReq)
	if err != nil {
		return fmt.Errorf("failed to marshal publish request: %v", err)
	}

	url := fmt.Sprintf("%s/api/publish", centrifugoConfig.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "apikey "+centrifugoConfig.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Centrifugo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Centrifugo API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func sendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Topic == "" || req.Content == "" || req.Author == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Topic, content, and author are required"})
		return
	}

	message := Message{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Topic:     req.Topic,
		Content:   req.Content,
		Author:    req.Author,
		Timestamp: time.Now(),
	}

	// Special handling for "all" channel: broadcast to all channels
	if req.Topic == "all" {
		// Get all available topics
		topics := []string{"all", "general", "tech", "random", "announcements"}
		
		// Publish to all topic channels
		for _, topic := range topics {
			channelName := fmt.Sprintf("topic:%s", topic)
			if err := publishToCentrifugo(channelName, message); err != nil {
				log.Printf("Failed to publish to '%s' topic: %v", topic, err)
				// Continue publishing to other channels even if one fails
			}
		}
	} else {
		// Normal behavior: publish to specific topic channel
		channelName := fmt.Sprintf("topic:%s", req.Topic)
		if err := publishToCentrifugo(channelName, message); err != nil {
			log.Printf("Failed to publish to Centrifugo: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
			return
		}

		// Also publish to the "all" topic channel to aggregate messages from all other topics
		allChannelName := "topic:all"
		if err := publishToCentrifugo(allChannelName, message); err != nil {
			log.Printf("Failed to publish to 'all' topic: %v", err)
			// Don't fail the request if publishing to "all" fails
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

func getTopics(c *gin.Context) {
	// For demo purposes, return some example topics
	// "all" is a special topic that aggregates messages from all other topics
	topics := []string{"all", "general", "tech", "random", "announcements"}
	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
		"centrifugo_url": centrifugoConfig.URL,
	})
}

func main() {
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// Routes
	r.GET("/health", healthCheck)
	r.GET("/api/topics", getTopics)
	r.POST("/api/messages", sendMessage)
	r.POST("/api/token", getToken)

	port := getEnv("PORT", "8080")
	log.Printf("Starting server on port %s", port)
	log.Printf("Centrifugo URL: %s", centrifugoConfig.URL)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}