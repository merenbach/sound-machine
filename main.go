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
// TODO: better playlist display in browser
// TODO: different rooms, namespaced to allow multiple "conversations"
// TODO: Slack integration
// TODO: dedicated client app to submit?
// TODO: allow refreshing of list if remote manifest updated??
// TODO: Ensure any necessary goroutine cleanup/waiting is performed.
// TODO: Add "release the kracken" sound
// TODO: Should we be creating sound/button elements in JS to dogfood the /sounds endpoint?
// TODO: use strings, not bytes, in client/hub; can we even eliminate need for hub?
// TODO: make client.send a string channel with size 0? 1? close when it's not draining? should it be the maximum number of clients?
// TODO: if centralizing broadcasting, can we allow specification of an event _and_ data, maybe with a custom message struct?
// TODO: IMPORTANT: explore client-to-server heartbeats to keep client registered

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

	router.GET("/sounds", func(c *gin.Context) {
		// TODO: potential race condition?
		c.JSON(http.StatusOK, sounds)
	})
	router.POST("/play/:sound", func(c *gin.Context) {
		sound := c.Param("sound")
		log.Println("Received:", sound)
		// TODO: potential race condition?
		if _, ok := sounds[sound]; ok {
			log.Println("Adding sound to queue:", sound)
			hub.Broadcast(sound)
			c.String(http.StatusOK, "")
		} else {
			log.Println("Ignoring unknown sound:", sound)
			c.String(http.StatusNotFound, "")
		}
	})
	router.GET("/play", func(c *gin.Context) {
		client := newClient()
		hub.Register(client)
		defer hub.Unregister(client)
		client.writePump(c)
	})
	SLACK_TOKEN := "hello"
	router.POST("/slack", func(c *gin.Context) {
		type SlackSlashCommand struct {
			Command     string `json:"command"`
			Text        string `json:"text"`
			UserName    string `json:"user_name"`
			ChannelName string `json:"channel_name"`
			Token       string `json:"token"`
		}
		var slashCommand SlackSlashCommand
		err := c.BindJSON(&slashCommand)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error: ", err.Error())
		}

		if token := slashCommand.Token; token == "" || token != SLACK_TOKEN {
			log.Println("Missing or invalid token: ", token, c.Request.PostForm)
			c.String(http.StatusUnauthorized, "")
			return
		}

		log.Println(slashCommand)

		// if command == '/play':
		// 	d['text'] = text.lower()
		// TODO: potential race condition?
		sound := strings.TrimSpace(slashCommand.Text)
		if _, ok := sounds[sound]; ok {
			hub.Broadcast(sound)
		} else {
			log.Println("Could not find a command")
			// TODO: potential race condition?
			c.JSON(http.StatusNotFound, "hello")
			return
		}

		// 	p = playlist()
		// 	# /play, /play nonsensaoeuthaeounthaeou, /play aoeu etuitituiti, etc.
		// 	if text not in p:
		// 		return '\n'.join(p)
		// elif command == '/say':
		// 	if not text:
		// 		return 'USAGE: /say [something]'
		// 	d['text'] = text
		// @lru_cache()
		// def response_type_for_channel(channel):
		// 	valid_channels = ['privategroup', 'directmessage']
		// 	valid_channels.extend(current_app.config['SLACK_CHANNELS'].split(','))
		// 	return 'in_channel' if channel in valid_channels else 'ephemeral'
		msg := fmt.Sprint(":speaker:", sound)
		respType := "in_channel"
		// resp_type = response_type_for_channel(channel_name)
		TXT_REPLIES := map[string]string{
			"56k": "56k response",
		}
		if slashCommand.Command == "/play" {
			if txtReply, ok := TXT_REPLIES[sound]; ok {
				// # [TODO] use 'pretext' instead?
				msg = fmt.Sprint(":speaker:", txtReply)
			} else {
				c.String(http.StatusOK, "")
				return
			}

			c.JSON(http.StatusOK, gin.H{"response_type": respType, "text": msg})
		}
	})
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// TODO: potential race condition?
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", struct {
			Sounds map[string]string
		}{
			Sounds: sounds,
		})
	})

	router.Run(":" + port)
}
