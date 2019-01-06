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

(function () {
    "use strict";
    window.onload = function () {

        /* ApplyHTMLCollection applies a function to an HTMLCollection.
         * There are quite a few ways to do this, including converting to an array.
         * For now, we just use a traditional method here.
         */
        function applyToHTMLCollection(els, f) {
            for (var i = 0; i < els.length; i++) {
                f(els[i]);
            }
        }

        // NewEventSource creates a new Server-Sent Event receiver for the given URL path.
        function newEventSource(urlPath) {
            var eventSource = new EventSource(urlPath);
            eventSource.onopen = () => console.log("EventSource connection to server opened:", urlPath);
            eventSource.onerror = () => console.error("EventSource failed:", urlPath);
            return eventSource;
        }

        // NewPlayer creates an HTML5 audio player from the given audio element mapping.
        function newPlayer(audioMap, trackEndedCallback) {
            console.log("Creating audio player with elements:", audioMap);
            const queue = [];
            (function playFromQueue() {
                setTimeout(function () {
                    if (queue.length == 0) {
                        playFromQueue();
                        return;
                    }

                    const next = queue.shift();
                    console.log("PLAY:", next);
                    const audio = audioMap[next];
                    // audio.onplay = () => currentTrack = audio;
                    audio.onended = () => {
                        playFromQueue();
                        trackEndedCallback();
                    }
                    audio.play();
                }, 100);
            }());
            return (sound) => queue.push(sound);
        }

        /*source.onmessage = function(e) {
          console.log("RECV: " + e.data);
          document.getElementById('charlie').innerHTML += e.data + '<br>';
        };*/
        // source.close();

        const audioElements = {};

        const sounds = document.getElementById("sounds");

        // Create central audio-name-to-element map.
        applyToHTMLCollection(sounds.getElementsByTagName("audio"), function (el) {
            audioElements[el.dataset.name] = el;
        });
        
        // Assign action to buttons.
        applyToHTMLCollection(sounds.getElementsByClassName("play"), function (el) {
            el.onclick = function (e) {
                e.preventDefault();
                const sound = el.dataset.sound;
                console.log("SEND:", sound);
                fetch("/play/" + sound, {
                    method: "POST"
                })
                    .catch((error) => console.error(error));
            };
        });
        
        const launch = document.getElementById("launch");
        launch.onclick = function () {
            launch.disabled = true;

            applyToHTMLCollection(sounds.getElementsByClassName("play"), function (el) {
                el.disabled = false;
            });

            const playlistElement = document.getElementById("playlist");
            const play = newPlayer(audioElements, () => playlistElement.children[0].remove());

            const source = newEventSource("/listen");
            source.addEventListener("message", function (e) {
                console.log("RECV:", e.data);
                play(e.data);

                var newElement = document.createElement("li");
                newElement.textContent = e.data;
                playlistElement.appendChild(newElement);
            });

            /*source.onmessage = function(e) {
              console.log("RECV: " + e.data);
              document.getElementById('charlie').innerHTML += e.data + '<br>';
            };*/
            // source.close();

            // fetch('/play/')
            //     .catch(function (e) {
            //         console.log(e);
            //     })
            //     .then(function (response) {
            //         if (response.ok) {
            //             return response.json();
            //         }
            //         throw new Error(response.statusText);
            //     })
            //     .catch(function (e) {
            //         console.log(e);
            //     })
            //     .then(function (obj) {
            //         const sounds = document.getElementById("sounds");
            //         Object.entries(obj).forEach(
            //             ([key, value]) => {
            //                 console.log(key, value);
            //                 const audio = new Audio(value);
            //                 audio.preload = 'auto';
            //                 sounds.appendChild(audio);

            //                 audioElements[key] = audio;

            //                 const button = document.createElement('a');
            //                 button.href = '#';
            //                 button.innerHTML = key;
            //                 sounds.appendChild(button);
            //                 button.onclick = function (event) {
            //                     event.preventDefault();
            //                     if (!source) {
            //                         return false;
            //                     }

            //                     console.log("SEND: " + key);
            //                     conn.send(key);
            //                     return false;
            //                 };
            //             }
            //         );
            //     });

        };
    };
}());