package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) request(method, endpoint string, body interface{}) ([]byte, error) {
	url := c.baseURL + endpoint
	var reqBody io.Reader

	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("error: %s", string(data))
	}

	return data, nil
}

func (c *Client) Login(email, password string) (*AuthResponse, error) {
	data, err := c.request("POST", "/api/v1/auth/login", AuthRequest{Email: email, Password: password})
	if err != nil {
		return nil, err
	}
	var resp AuthResponse
	json.Unmarshal(data, &resp)
	return &resp, nil
}

func (c *Client) Register(email, password string) (*AuthResponse, error) {
	data, err := c.request("POST", "/api/v1/auth/register", AuthRequest{Email: email, Password: password})
	if err != nil {
		return nil, err
	}
	var resp AuthResponse
	json.Unmarshal(data, &resp)
	return &resp, nil
}

func (c *Client) SearchSpotify(query string) ([]Track, error) {
	data, err := c.request("GET", fmt.Sprintf("/api/v1/search/spotify?q=%s", query), nil)
	if err != nil {
		return nil, err
	}
	var tracks []Track
	json.Unmarshal(data, &tracks)
	return tracks, nil
}

func (c *Client) SearchSoundCloud(query string) ([]Track, error) {
	data, err := c.request("GET", fmt.Sprintf("/api/v1/search/soundcloud?q=%s", query), nil)
	if err != nil {
		return nil, err
	}
	var tracks []Track
	json.Unmarshal(data, &tracks)
	return tracks, nil
}

func (c *Client) SearchYouTube(query string) ([]Track, error) {
	data, err := c.request("GET", fmt.Sprintf("/api/v1/search/youtube?q=%s", query), nil)
	if err != nil {
		return nil, err
	}
	var tracks []Track
	json.Unmarshal(data, &tracks)
	return tracks, nil
}

func (c *Client) GetStreamURL(source, sourceID string) (string, error) {
	switch source {
	case "soundcloud":
		data, err := c.request("GET", "/api/v1/stream/soundcloud/"+sourceID, nil)
		if err != nil {
			return "", err
		}
		var resp struct {
			URL       string `json:"url"`
			Format    string `json:"format"`
			ExpiresAt int64  `json:"expires_at"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return "", err
		}
		return resp.URL, nil
	case "youtube":
		data, err := c.request("GET", "/api/v1/stream/youtube/"+sourceID, nil)
		if err != nil {
			return "", err
		}
		var resp struct {
			URL       string `json:"url"`
			Format    string `json:"format"`
			ExpiresAt int64  `json:"expires_at"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return "", err
		}
		return resp.URL, nil
	default:
		return "", fmt.Errorf("stream não implementado para source=%s", source)
	}
}

func (c *Client) GetProviders() ([]string, error) {
	data, err := c.request("GET", "/api/v1/providers", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Providers []string `json:"providers"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Providers, nil
}

func (c *Client) GetPlaylists() ([]Playlist, error) {
	data, err := c.request("GET", "/api/v1/playlists", nil)
	if err != nil {
		return nil, err
	}
	var playlists []Playlist
	json.Unmarshal(data, &playlists)
	return playlists, nil
}

func (c *Client) CreatePlaylist(name string) (*Playlist, error) {
	data, err := c.request("POST", "/api/v1/playlists", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var playlist Playlist
	json.Unmarshal(data, &playlist)
	return &playlist, nil
}

func (c *Client) AddTrackToPlaylist(playlistID uint, track Track) error {
	_, err := c.request("POST", fmt.Sprintf("/api/v1/playlists/%d/tracks", playlistID), track)
	return err
}

func (c *Client) GetSpotifyStatus() (bool, error) {
	data, err := c.request("GET", "/api/v1/auth/spotify/status", nil)
	if err != nil {
		return false, err
	}

	var resp map[string]bool
	json.Unmarshal(data, &resp)
	return resp["linked"], nil
}

func (c *Client) GetSpotifyAuthURL() (string, error) {
	data, err := c.request("GET", "/api/v1/auth/spotify", nil)
	if err != nil {
		return "", err
	}

	var resp map[string]string
	json.Unmarshal(data, &resp)
	return resp["url"], nil
}

func (c *Client) PlayTrack(uri, deviceID string) error {
	body := map[string]string{"uri": uri}
	if deviceID != "" {
		body["device_id"] = deviceID
	}
	_, err := c.request("POST", "/api/v1/player/play", body)
	return err
}

func (c *Client) Resume() error {
	_, err := c.request("POST", "/api/v1/player/play", nil)
	return err
}

func (c *Client) Pause() error {
	_, err := c.request("POST", "/api/v1/player/pause", nil)
	return err
}

func (c *Client) Next() error {
	_, err := c.request("POST", "/api/v1/player/next", nil)
	return err
}

func (c *Client) Previous() error {
	_, err := c.request("POST", "/api/v1/player/previous", nil)
	return err
}

func (c *Client) Devices() ([]map[string]interface{}, error) {
	data, err := c.request("GET", "/api/v1/player/devices", nil)
	if err != nil {
		return nil, err
	}
	var devs []map[string]interface{}
	_ = json.Unmarshal(data, &devs)
	return devs, nil
}
