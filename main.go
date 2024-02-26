// main.go

package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/cors"

	// "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

var rdb *redis.Client

// User struct represents a user
type User struct {
	Username string `json:"username"`
	Points   int    `json:"points"`
}

// Initialize Redis client
func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password
		DB:       0,                // Default DB
	})
}

// GetAllUsers returns all users
// GetAllUsers returns all users from Redis
// GetAllUsers returns all users from Redis
func GetAllUsers(c *gin.Context) {
	keys, err := rdb.Keys(ctx, "user:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user keys from Redis", "details": err.Error()})
		return
	}

	var users []User
	for _, key := range keys {
		userData, err := rdb.Get(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data from Redis", "key": key, "details": err.Error()})
			return
		}

		var user User
		err = json.Unmarshal([]byte(userData), &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unmarshal user data", "key": key, "details": err.Error()})
			return
		}

		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// GetUserByName returns a user by username
func GetUserByName(c *gin.Context) {
	username := c.Param("username")

	// Fetch user data from Redis
	userData, err := rdb.Get(ctx, "user:"+username).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var user User
	err = json.Unmarshal([]byte(userData), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// PutUser creates a new user
func PutUser(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Check if the user already exists
	_, err := rdb.Get(ctx, "user:"+newUser.Username).Result()
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	// Add the new user to Redis
	userJSON, err := json.Marshal(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	err = rdb.Set(ctx, "user:"+newUser.Username, userJSON, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, newUser)
}

// UpdatePoints updates user points by username
func UpdatePoints(c *gin.Context) {
	username := c.Param("username")

	var updateData struct {
		Points int `json:"points"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Fetch current user data from Redis
	userData, err := rdb.Get(ctx, "user:"+username).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var currentUser User
	err = json.Unmarshal([]byte(userData), &currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Update points and save back to Redis
	currentUser.Points = updateData.Points
	userJSON, err := json.Marshal(currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	err = rdb.Set(ctx, "user:"+currentUser.Username, userJSON, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, currentUser)
}

// Your other handlers and main function remain the same

func main() {
	r := gin.Default()

	// Use CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Add your React app's URL
	r.Use(cors.New(config))

	r.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Deployed",
		})
	})

	r.GET("/users", GetAllUsers)
	r.GET("/users/:username", GetUserByName)
	r.POST("/users", PutUser)
	r.PUT("/users/:username", UpdatePoints)

	r.Run(":8080")
}
