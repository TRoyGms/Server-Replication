package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"serverreplication/structs"
)

var (
	users      = []structs.User{}
	usersMutex = sync.Mutex{}
	nextID     = 1
	newChanges = false
)

func main() {
	r := gin.Default()

	r.GET("/users", getUsers)
	r.POST("/users", createUser)
	r.PUT("/users/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)
	r.GET("/users/check-new", checkNewRecords)

	log.Println("Principal server is running on port 8080")
	log.Fatal(r.Run(":8080"))
}

func getUsers(c *gin.Context) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	c.JSON(http.StatusOK, users)
}

func createUser(c *gin.Context) {
	var newUser structs.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	newUser.ID = nextID
	nextID++
	users = append(users, newUser)
	newChanges = true

	c.JSON(http.StatusCreated, newUser)
}

func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedUser structs.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	for i, user := range users {
		if user.ID == id {
			users[i].Name = updatedUser.Name
			users[i].Username = updatedUser.Username
			newChanges = true
			c.JSON(http.StatusOK, users[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			newChanges = true
			c.Status(http.StatusNoContent)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func checkNewRecords(c *gin.Context) {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"newChanges": newChanges})
	newChanges = false
}
