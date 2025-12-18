package comm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type images struct {
	Url string `json:"url"`
	X   int    `json:"x"`
	Y   int    `json:"y"`
	W   int    `json:"w"`
	H   int    `json:"h"`
}
type lines struct {
	Text  string `json:"text"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
	Size  int    `json:"size"`
	Width int    `json:"width"`
}
type qrcodes struct {
	Text string `json:"text"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Size int    `json:"size"`
}
type blocks struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	H      int    `json:"h"`
	Color  string `json:"color"`
	Radius int    `json:"radius"`
}
type sendBody struct {
	BgColor string    `json:"bg_color"`
	Width   int       `json:"width"`
	Height  int       `json:"height"`
	Quality int       `json:"quality"`
	Format  string    `json:"format"`
	Author  string    `json:"author"`
	Images  []images  `json:"images"`
	Lines   []lines   `json:"lines"`
	Qrcodes []qrcodes `json:"qrcodes"`
	Blocks  []blocks  `json:"blocks"`
}
type ReqData struct {
	Url      string `json:"url"`
	ExpireAt string `json:"expireAt"`
}
type Req struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    ReqData `json:"data"`
}

func GeneratePoster(route string, teamName string, teamSlogan string, countOfMember int, teamMember []string) (string, error) {
	var sendB sendBody
	sendB.BgColor = "#ffffff"
	sendB.Width = 1080
	sendB.Height = 1920
	sendB.Quality = 80
	sendB.Format = "jpg"
	sendB.Author = "zjutjh"

	// Background image
	sendB.Images = append(sendB.Images, images{
		Url: "https://walk.zjutjh.com/file/poster_bg.jpg", // Assuming this URL is valid or needs config
		X:   0,
		Y:   0,
		W:   1080,
		H:   1920,
	})

	// Team Name
	sendB.Lines = append(sendB.Lines, lines{
		Text:  teamName,
		X:     540,
		Y:     400,
		Color: "#000000",
		Size:  60,
		Width: 800,
	})

	// Route
	sendB.Lines = append(sendB.Lines, lines{
		Text:  route,
		X:     540,
		Y:     500,
		Color: "#333333",
		Size:  40,
		Width: 800,
	})

	// Slogan
	sendB.Lines = append(sendB.Lines, lines{
		Text:  teamSlogan,
		X:     540,
		Y:     600,
		Color: "#666666",
		Size:  30,
		Width: 800,
	})

	// Members
	y := 800
	for _, member := range teamMember {
		sendB.Lines = append(sendB.Lines, lines{
			Text:  member,
			X:     540,
			Y:     y,
			Color: "#000000",
			Size:  35,
			Width: 800,
		})
		y += 50
	}

	sending, err := json.Marshal(sendB)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(sending)
	request, _ := http.NewRequest("POST", "https://api.imgrender.cn/open/v1/pics", reader)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	// API Key should be in config, but for now hardcoding as in main branch or placeholder
	request.Header.Set("X-API-Key", "171123855488848778.iuHQWhIvOjJqIVmKDSvXj1pWfegMoePk")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)

	var dReq Req
	err = json.Unmarshal(body, &dReq)
	if err != nil {
		return "", err
	}

	if dReq.Code != 200 {
		return "", fmt.Errorf("poster api error: %s", dReq.Message)
	}

	return dReq.Data.Url, nil
}
