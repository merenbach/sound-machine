// Copyright 2018 Andrew Merenbach
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// TODO: emoji responses? handles for participants?
// TODO: revamp fault tolerance (invalid sound, etc.)
// TODO: better log/history display in browser, plus status messages about joins/leaves--and don't try to play those...
// TODO: different rooms, namespaced to allow multiple "conversations"
// TODO: Slack integration
// TODO: dedicated client app to submit?
// NOTE: portions based heavily on https://github.com/gorilla/websocket/tree/master/examples/chat
// TODO: allow refreshing of list if remote manifest updated??
// TODO: remove trailing /ws if we can switch to Heroku for POC
// TODO: Ensure any necessary goroutine cleanup/waiting is performed.
// TODO: Add "release the kracken" sound

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

// LoadStringMap loads a string map from the given path.
func loadStringMap(p string) map[string]string {
	bb, err := ioutil.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}

	var m map[string]string
	if err := json.Unmarshal(bb, &m); err != nil {
		log.Fatal(err)
	}
	return m
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	hub := newHub()
	go hub.run()

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	sounds := loadStringMap("sounds.json")
	log.Println("Loading sounds:", sounds)
	//playlist := make(chan string, 100)
	//defer close(playlist)

	/*currentSound := ""
	go func() {
		select {
		case ch := <-playlist:
			// new sound, and if we're here we're ready
			currentSound = ch
			time.Sleep(1 * time.Second)
		default:
			time.Sleep(5 * time.Second)
		}
		time.Sleep(1 * time.Second)
	}()*/
	/*go func(r chan<- string, s <-chan string) {

	}(queueSicle, drainSicle)*/

	router.GET("/sounds", func(c *gin.Context) {
		c.JSON(http.StatusOK, sounds)
	})
	router.POST("/play/:sound", func(c *gin.Context) {
		sound := c.Param("sound")
		log.Println("Received:", sound)
		if _, ok := sounds[sound]; ok {
			log.Println("Adding sound to queue:", sound)
			hub.broadcast <- []byte(sound)
			c.String(http.StatusOK, "")
		} else {
			log.Println("Ignoring unknown sound:", sound)
			c.String(http.StatusNotFound, "")
		}
	})
	router.GET("/listen", func(c *gin.Context) {
		registerClient(hub, c)
	})
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", struct {
			Sounds map[string]string
		}{
			Sounds: sounds,
		})
	})

	router.Run(":" + port)
}
