package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"serverreplication/structs"
)

var (
	replicatedUsers = []structs.User{}
	replicaMutex    = sync.Mutex{}
	server1URL      = "http://localhost:8080/users"
	checkURL        = "http://localhost:8080/users/check-new"
)

func main() {
	go startReplication()

	r := gin.Default()

	r.GET("/replication/short", shortRoute)
	r.GET("/replication/long", longRoute)
	r.GET("/replication/data", getReplicatedUsers)

	log.Println("Replication server is running on port 8081")
	log.Fatal(r.Run(":8081"))
}

func startReplication() {
	for {
		log.Println("Executing short polling...")
		time.Sleep(5 * time.Second)

		// Hacer la solicitud de short polling
		response, err := http.Get(checkURL)
		if err != nil {
			log.Println("Error during short polling:", err)
			continue
		}

		var result map[string]bool
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			log.Println("Error decoding short polling response:", err)
			response.Body.Close() // Cerrar explícitamente el cuerpo
			continue
		}
		response.Body.Close() // Cerrar explícitamente el cuerpo aquí también

		// Si hay nuevos cambios, ejecutar el long polling
		if result["newChanges"] {
			log.Println("Changes detected! Executing long polling...")
			if err := performLongPolling(); err != nil {
				log.Println("Error during long polling:", err)
			} else {
				log.Println("Data successfully synchronized.")
			}
		} else {
			log.Println("No changes detected.")
		}
	}
}


func performLongPolling() error {
	log.Println("Fetching user data from Server 1...")
	response, err := http.Get(server1URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var users []structs.User
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		return err
	}

	replicaMutex.Lock()
	replicatedUsers = users
	replicaMutex.Unlock()

	log.Println("User data successfully updated.")
	return nil
}

func shortRoute(c *gin.Context) {
	log.Println("Short polling endpoint accessed manually.")
	response, err := http.Get(checkURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for new records"})
		return
	}
	defer response.Body.Close()

	var result map[string]bool
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func longRoute(c *gin.Context) {
	log.Println("Long polling endpoint accessed manually.")
	err := performLongPolling()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, replicatedUsers)
}

func getReplicatedUsers(c *gin.Context) {
	replicaMutex.Lock()
	defer replicaMutex.Unlock()

	c.JSON(http.StatusOK, replicatedUsers)
}
