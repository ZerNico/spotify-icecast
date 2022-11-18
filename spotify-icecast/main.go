package main

// import gofiber
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"os"
	"os/exec"
)

// event struct
type Event struct {
	PlayerEvent string `json:"player_event"`
	TrackID     string `json:"track_id"`
	OldTrackID  string `json:"old_track_id"`
}

type Track struct {
	Artists []struct {
		Name string `json:"name"`
	} `json:"artists"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

var token = ""
var darkice *exec.Cmd

func main() {
	// create new fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Post("/event", func(c *fiber.Ctx) error {
		event := new(Event)

		if err := c.BodyParser(&event); err != nil {
			return err
		}

		if event.PlayerEvent == "playing" {
			handlePlaying(*event)
		}

		if event.PlayerEvent == "session_connected" {
			handleSessionConnected(*event)
		}

		if event.PlayerEvent == "session_disconnected" {
			handleSessionDisconnected(*event)
		}

		return c.SendString("OK")
	})

	// start server
	app.Listen(":8080")
}

func handleSessionConnected(event Event) {
	if darkice != nil {
		handleSessionDisconnected(event)
	}

	args := []string{"-c", "./darkice.cfg"}
	darkice = exec.Command("darkice", args...)

	if err := darkice.Start(); err != nil {
		fmt.Println("failed to start process: ", err)
	}
}

func handleSessionDisconnected(event Event) {
	if darkice == nil {
		return
	}
	if err := darkice.Process.Kill(); err != nil {
		fmt.Println("failed to kill process: ", err)
	}
	darkice = nil
}

func handlePlaying(event Event) {
	if token == "" {
		updateToken()
	}

	var track Track
	var err1 error

	track, err1 = getTrack(event.TrackID)
	if err1 != nil {
		var err2 error
		updateToken()
		track, err2 = getTrack(event.TrackID)
		if err2 != nil {
			return
		}
	}

	updateMetadata(track)
}

func getTrack(id string) (Track, error) {

	var track Track

	var bearer = "Bearer " + token

	req, _ := http.NewRequest("GET", "https://api.spotify.com/v1/tracks/"+id, nil)
	req.Header.Add("Authorization", bearer)
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return track, err
	}
	defer resp.Body.Close()

	// if status code is 401, token is invalid
	if resp.StatusCode == 401 {
		return track, fmt.Errorf("invalid token")
	}

	body, err := io.ReadAll(resp.Body)

	if err := json.Unmarshal(body, &track); err != nil {
		return track, err
	}

	return track, nil
}

func updateMetadata(track Track) {
	// song name with title - artist, artist2 etc.
	var song = ""
	for i, artist := range track.Artists {
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

func updateToken() {
	args := []string{os.Getenv("SPOTIFY_USERNAME"), os.Getenv("SPOTIFY_PASSWORD")}
	out, err := exec.Command("/home/user/librespot-token", args...).Output()
	if err != nil {
		fmt.Println(err)
	}
	newToken := string(out)
	// remove trailing newline
	token = newToken[:len(newToken)-1]
}
