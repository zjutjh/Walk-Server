package utility

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type texts struct {
	X           int     `json:"x"`
	Y           int     `json:"y"`
	Text        string  `json:"text"`
	Width       int     `json:"width"`
	Font        string  `json:"font"`
	FontSize    int     `json:"fontSize"`
	LineHeight  int     `json:"lineHeight"`
	LineSpacing float32 `json:"lineSpacing"`
	Color       string  `json:"color"`
	TextAlign   string  `json:"textAlign"`
	ZIndex      int     `json:"zIndex"`
}
type images struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	ZIndex int    `json:"zIndex"`
}
type lines struct {
	StartX int    `json:"startX"`
	StartY int    `json:"startY"`
	EndX   int    `json:"endX"`
	EndY   int    `json:"endY"`
	Width  int    `json:"width"`
	Color  string `json:"color"`
	ZIndex int    `json:"zIndex"`
}
type qrcodes struct {
	X               int    `json:"x"`
	Y               int    `json:"y"`
	Size            int    `json:"size"`
	Content         string `json:"content"`
	ForegroundColor string `json:"foregroundColor"`
	BackgroundColor string `json:"backgroundColor"`
	ZIndex          int    `json:"zIndex"`
}
type blocks struct {
	X               int    `json:"x"`
	Y               int    `json:"y"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	BackgroundColor string `json:"backgroundColor"`
	ZIndex          int    `json:"zIndex"`
}
type Send struct {
	Width           int       `json:"width"`
	Height          int       `json:"height"`
	BackgroundColor string    `json:"backgroundColor"`
	Texts           []texts   `json:"texts"`
	Images          []images  `json:"images"`
	Lines           []lines   `json:"lines"`
	Qrcodes         []qrcodes `json:"qrcodes"`
	Blocks          []blocks  `json:"blocks"`
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

func Pic(Route string, TeamName string, TeamSlogan string, CountOfMember int, TeamMember []string) (string, error) {
	var ItemCount int
	var PicUrl string
	ItemCount = 0
	PicUrl = "http://bilibili12433014.top:8088/chfs/shared/Downloads/1/"
	sendB := Send{
		Width:           767,
		Height:          1085,
		BackgroundColor: "#ffffff",
	}
	sendB.Images = append(sendB.Images, images{
		X:      0,
		Y:      0,
		Url:    PicUrl + Route + ".jpg",
		Width:  767,
		Height: 1085,
		ZIndex: ItemCount,
	})
	ItemCount++
	var NameString string
	for i := 0; i < CountOfMember; i++ {
		NameString += TeamMember[i] + "\n"
	}
	var Color string
	if strings.Contains(Route, "朝晖") {
		Color = "#331B14"
	}
	if strings.Contains(Route, "屏峰") {
		Color = "#1C1C42"
	}
	if strings.Contains(Route, "莫干山") {
		Color = "#243A24"
	}
	sendB.Texts = append(sendB.Texts, texts{
		X:           383,
		Y:           300,
		Text:        NameString,
		Width:       767,
		Font:        "Alibaba-PuHuiTi-Heavy",
		FontSize:    90,
		LineHeight:  90,
		LineSpacing: 1.2,
		Color:       Color,
		TextAlign:   "center",
		ZIndex:      ItemCount,
	})
	ItemCount++
	sendB.Images = append(sendB.Images, images{
		X:      0,
		Y:      0,
		Url:    PicUrl + Route + ".png",
		Width:  767,
		Height: 1085,
		ZIndex: ItemCount,
	})
	ItemCount++
	sendB.Texts = append(sendB.Texts, texts{
		X:           30,
		Y:           15,
		Text:        TeamName,
		Width:       767,
		Font:        "Alibaba-PuHuiTi-Heavy",
		FontSize:    70,
		LineHeight:  70,
		LineSpacing: 1,
		Color:       "#FFFFFF",
		TextAlign:   "left",
		ZIndex:      ItemCount,
	})
	ItemCount++
	sendB.Texts = append(sendB.Texts, texts{
		X:           740,
		Y:           100,
		Text:        TeamSlogan,
		Width:       767,
		Font:        "Alibaba-PuHuiTi-Heavy",
		FontSize:    25,
		LineHeight:  25,
		LineSpacing: 1,
		Color:       "#DDDDDD",
		TextAlign:   "right",
		ZIndex:      ItemCount,
	})
	ItemCount++
	sending, err := json.Marshal(sendB)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(sending)
	request, _ := http.NewRequest("POST", "https://api.imgrender.cn/open/v1/pics", reader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "189950.Ntr3pucE6vfg02xMZU3tgMZCFaWLD+h4sUUgeMxNpWo=")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	var dReq Req
	err = json.Unmarshal(body, &dReq)
	if err != nil {
		return "", err
	}
	fmt.Println(dReq)
	if dReq.Code != 0 {
		return "", errors.New(dReq.Message)
	} else {
		return dReq.Data.Url, nil
	}
}
