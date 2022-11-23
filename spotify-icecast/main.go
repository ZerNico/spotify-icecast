package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

type Event struct {
	Event string `json:"event"`
	Track Track  `json:"track"`
}

type Track struct {
	Artist []struct {
		Name string `json:"name"`
	} `json:"artist"`
	Name string `json:"name"`
}

var darkice *exec.Cmd

func main() {
	messageOut := make(chan string)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := url.URL{Scheme: "ws", Host: "localhost:24879", Path: "/events"}
	log.Print("connecting to librespot")

	var c *websocket.Conn

	for {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

		if err != nil {
			log.Printf("handshake failed")
			time.Sleep(1 * time.Second)
		} else {
			c = conn
			break
		}
  }

	//When the program closes, close the connection
	defer c.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			var event Event
			err = json.Unmarshal(message, &event)
			if err != nil {
				log.Println("Unmarshal:", err)
				return
			}

			if event.Event == "metadataAvailable" {
				handleMetadata(event.Track)
			}

			if event.Event == "inactiveSession" {
				handleSessionDisconnected()
			}

			if event.Event == "contextChanged" {
				handleSessionConnected()
			}
		}

	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case m := <-messageOut:
			log.Printf("Send Message %s", m)
			err := c.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func handleSessionConnected() {
	log.Println("Session Connected")
	if darkice != nil {
		return
	}

	args := []string{"-c", "./darkice.cfg"}
	darkice = exec.Command("darkice", args...)

	if err := darkice.Start(); err != nil {
		fmt.Println("failed to start process: ", err)
		darkice = nil
	}
}

func handleSessionDisconnected() {
	log.Println("Session Disconnected")
	if darkice == nil {
		return
	}
	if err := darkice.Process.Kill(); err != nil {
		fmt.Println("failed to kill process: ", err)
	}
	darkice = nil
}

func handleMetadata(track Track) {
	// song name with title - artist, artist2 etc.
	var song = ""
	for i, artist := range track.Artist {
		if i > 0 {
			song += ", "
		}
		song += artist.Name
	}
	song += " - " + track.Name

	isHTTPS := os.Getenv("ICECAST2_HTTPS")

	icecastURL := ""
	if isHTTPS == "true" {
		icecastURL = "https://"
	} else {
		icecastURL = "http://"
	}
	icecastURL += os.Getenv("ICECAST2_IP") + ":" + os.Getenv("ICECAST2_PORT") + "/admin/metadata"

	req, _ := http.NewRequest("GET", icecastURL, nil)
	query := req.URL.Query()
	query.Add("mode", "updinfo")
	query.Add("song", song)
	query.Add("description", os.Getenv("ICECAST2_DESCRIPTION"))
	query.Add("name", os.Getenv("ICECAST2_NAME"))
	query.Add("mount", "/"+os.Getenv("ICECAST2_MOUNT"))
	req.URL.RawQuery = query.Encode()

	req.Header.Add("Authorization", "Basic "+basicAuth("source", os.Getenv("ICECAST2_PASSWORD")))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
