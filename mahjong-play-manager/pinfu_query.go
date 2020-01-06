package main

import (
	"log"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"math"
)

const (
	windEastOfPinfuQuery = 27
)

type PinfuQuery struct {
	Man string `json:"man"`
	Pin string `json:"pin"`
	Sou string `json:"sou"`
	Honors string `json:"honors"`
	PlayerWind int `json:"player_wind"`
	RoundWind int `json:"round_wind"`
	WinTileType string `json:"win_tile_type"`
	WinTileValue string `json:"win_tile_value"`
}

type PinfuInfo struct {
	IsPinfu bool
	Cost int
}

func (p *PinfuQuery) Query() *PinfuInfo {
	url := "http://host.docker.internal:8000"
	j, _ := json.Marshal(p)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("response body:%s", body)
	pinfuInfo := PinfuInfo {}
	json.Unmarshal(body, &pinfuInfo)
	return &pinfuInfo
}

func (p *PinfuQuery) Parse(hands []int, ronTileId int, playerWind int, roundWind int) {
	hands = append(hands, ronTileId)
	for _, tileId := range hands {
		switch {
		case tileId < 36:
			p.Man += strconv.Itoa(int(math.Floor(float64(tileId/4))) + 1)
		case tileId < 72:
			p.Pin += strconv.Itoa(int(math.Floor(float64((tileId - 36)/4))) + 1)
		case tileId < 108:
			p.Sou += strconv.Itoa(int(math.Floor(float64((tileId - 72)/4))) + 1)
		case tileId < 136:
			p.Honors += strconv.Itoa(int(math.Floor(float64((tileId - 108)/4))) + 1)
		}
	}

	p.PlayerWind = p.WindForPinfuQuery(playerWind)
	p.RoundWind = p.WindForPinfuQuery(roundWind)
	switch {
	case ronTileId < 36:
		p.WinTileType = "man"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64(ronTileId/4))) + 1)
	case ronTileId < 72:
		p.WinTileType = "pin"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64((ronTileId - 36)/4))) + 1)
	case ronTileId < 108:
		p.WinTileType = "sou"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64((ronTileId - 72)/4))) + 1)
	case ronTileId < 136:
		p.WinTileType = "honors"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64((ronTileId - 108)/4))) + 1)
	}
	if len(p.Sou) == 14 && p.WinTileType == "sou" {
		p.Man = p.Sou
		p.Sou = ""
		p.WinTileType = "man"
	}
}

func (p *PinfuQuery) WindForPinfuQuery(wind int) int {
	return wind + windEastOfPinfuQuery - 1
}

/*
func main() {
	p := PinfuQuery{}
	p.Parse([]int{0,1,12,16,20,24,28,32,60,64,68,76,80}, 84, 27, 27)
	log.Println(p)
	p.Query()
}
*/