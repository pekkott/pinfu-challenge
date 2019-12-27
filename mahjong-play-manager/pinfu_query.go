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

type PinfuResponse struct {
	IsPinfu bool
}

func (p *PinfuQuery) Query() bool {
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
	pinfuResponse := PinfuResponse {}
	json.Unmarshal(body, &pinfuResponse)
	return pinfuResponse.IsPinfu
}

func (p *PinfuQuery) Parse(hands []int, ronHaiId int, playerWind int, roundWind int) {
	hands = append(hands, ronHaiId)
	for _, haiId := range hands {
		switch {
		case haiId < 36:
			p.Man += strconv.Itoa(int(math.Floor(float64(haiId/4))) + 1)
		case haiId < 72:
			p.Pin += strconv.Itoa(int(math.Floor(float64((haiId - 36)/4))) + 1)
		case haiId < 108:
			p.Sou += strconv.Itoa(int(math.Floor(float64((haiId - 72)/4))) + 1)
		case haiId < 136:
			p.Honors += strconv.Itoa(int(math.Floor(float64((haiId - 108)/4))) + 1)
		}
	}

	p.PlayerWind = playerWind
	p.RoundWind = roundWind
	switch {
	case ronHaiId < 36:
		p.WinTileType = "man"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64(ronHaiId/4))) + 1)
	case ronHaiId < 72:
		p.WinTileType = "pin"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64((ronHaiId - 36)/4))) + 1)
	case ronHaiId < 108:
		p.WinTileType = "sou"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64((ronHaiId - 72)/4))) + 1)
	case ronHaiId < 136:
		p.WinTileType = "honors"
		p.WinTileValue = strconv.Itoa(int(math.Floor(float64((ronHaiId - 108)/4))) + 1)
	}
}
/*
func main() {
	p := PinfuQuery{}
        p.Parse([]int{0,1,12,16,20,24,28,32,60,64,68,76,80}, 84, 27, 27)
        log.Println(p)
	p.Query()
}
*/