package main

import (
	"log"
	"sort"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"math"
	"math/rand"
	"encoding/json"
	"sync"
)

const (
	haiInMountNumber = 136
	haiInHandNumber = 13
	playerNumber = 4
	roundNumber = 4
	pointStart = 25000
)

var umaByOrder = [...]int{20, 10, -10, -20}

type Wind int

type MahjongPlayManager struct {
	round *Round
	playerNumber int
	playerIdInTurn int
	playerInfos []*PlayerInfo
	mount []int
	mountPosition int
	waitingNext bool
	waitingNextMux sync.Mutex
	sendMessages []*SendMessage
}

type PlayInfo struct {
	Round *Round `json:"round"`
	PlayerInfo *PlayerInfo `json:"playerInfo"`
	PlayerIds []int `json:"playerIds"`
	Winds []Wind `json:"winds"`
	Points []int `json:"points"`
}

type PlayerInfo struct {
	PlayerId int `json:"playerId"`
	Point int `json:"-"`
	FirstPinfuOrder int `json:"-"`
	Wind Wind `json:"wind"`
	Hands []int `json:"hands"`
	DrawnHai int `json:"drawnHai"`
	ReleasedHaiUp int `json:"releasedHaiUp"`
	CanRon bool `json:"canRon"`
}

type ReleasedHaiInfo struct {
	PlayerPosition int `json:"playerPosition"`
	ReleasedHai int `json:"releasedHai"`
	CanRon bool `json:"canRon"`
}

type RonInfo struct {
	Point int `json:"point"`
	PointsDiff int `json:"pointDiff"`
}

type Round struct {
	Wind Wind `json:"wind"`
        Round int `json:"round"`
        SubRound int `json:"subRound"`
}

type Result struct {
        Point int `json:"point"`
        Order int `json:"order"`
}

type SendMessage struct {
	Type string `json:"type"`
        Values interface{} `json:"values"`
}

const (
	EAST Wind = iota + 1
	SOUTH
	WEST
	NORTH
)

func WindList() []Wind {
	return []Wind{EAST, SOUTH, WEST, NORTH}
}

func (m *MahjongPlayManager) Init() {
	m.round = &Round{EAST, 1, 0}
	m.playerNumber = -1
	m.playerInfos = make([]*PlayerInfo, playerNumber)
	for i := range m.playerInfos {
		m.playerInfos[i] = &PlayerInfo{i, pointStart, playerNumber + 1, WindList()[i], make([]int, haiInHandNumber), -1, -1, false}
	}
	m.playerInfos[0].Point = 26000
	m.playerInfos[2].Point = 24000
	m.InitPlayerIdInTrun()
	m.InitSeed()
	m.waitingNext = false
	m.sendMessages = make([]*SendMessage, playerNumber)
}

func (m *MahjongPlayManager) InitPlayerIdInTrun() {
	for _, p := range m.playerInfos {
		if p.Wind == EAST {
			m.playerIdInTurn = p.PlayerId
			break
		}
	}
}

func (m *MahjongPlayManager) InitSeed() {
	var b [8]byte
	crypto_rand.Read(b[:])
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func (m *MahjongPlayManager) InitGame() {
	m.round = &Round{EAST, 1, 0}
	m.InitRound()
}

func (m *MahjongPlayManager) InitRound() {
        m.mountPosition = 0
	m.InitMount()
	m.InitHands()
	m.DistributeHai()
}

func (m *MahjongPlayManager) InitMount() {
	m.mount = make([]int, haiInMountNumber)
	for i := range m.mount {
		m.mount[i] = i
	}
	rand.Shuffle(len(m.mount), func(i, j int) {
		m.mount[i], m.mount[j] = m.mount[j], m.mount[i]
	})
	log.Println(m.mount)
}

func (m *MahjongPlayManager) InitHands() {
	for i := 0; i < haiInHandNumber; i++ {
		for j := 0; j < playerNumber; j++ {
			m.playerInfos[j].Hands[i] = m.mount[m.mountPosition]
			m.mountPosition++
		}
	}

	for i := 0; i < playerNumber; i++ {
		sort.Ints(m.playerInfos[i].Hands)
	}
}

func (m *MahjongPlayManager) newPlayerNumber() int {
	m.playerNumber++
	m.playerNumber = m.playerNumber % playerNumber
	return m.playerNumber
}

func (m *MahjongPlayManager) isReady() bool {
	log.Printf("playerNumber:%d", m.playerNumber)
	return m.playerNumber == playerNumber - 1
}

func (m *MahjongPlayManager) GenerateEachPlayerIds(playerId int) []int {
	playerIds := make([]int, playerNumber)
	for i := range playerIds {
		playerIds[i] = (playerNumber + playerId - i) % playerNumber
	}
	return playerIds
}

func (m *MahjongPlayManager) GenerateEachWinds(playerId int) []Wind {
	winds := make([]Wind, playerNumber)
	for i, p := range m.playerInfos {
		winds[(playerNumber - playerId + i) % playerNumber] = p.Wind
	}
	return winds
}

func (m *MahjongPlayManager) GeneratePoints() []int {
	points := make([]int, playerNumber)
	for _, p := range m.playerInfos {
		points[p.PlayerId] = p.Point
	}
	return points
}

func (m *MahjongPlayManager) DistributeHai() {
	log.Printf("mount position:%d", m.mountPosition);
	if m.mountPosition < haiInMountNumber {
		m.playerInfos[m.playerIdInTurn].DrawnHai = m.mount[m.mountPosition]
		m.mountPosition++
	}
}

func (m *MahjongPlayManager) ReleaseHai(position int) int {
	playerInTurn := m.playerInfos[m.playerIdInTurn]
	releasedHai := playerInTurn.DrawnHai
	if position >= 0 && position < len(playerInTurn.Hands) {
		log.Printf("release position:%d", position)
		releasedHai = playerInTurn.Hands[position]
		playerInTurn.Hands[position] = playerInTurn.DrawnHai
		sort.Ints(playerInTurn.Hands)
	}
	log.Printf("release drawnHai:%d", playerInTurn.DrawnHai)
	log.Printf("releasedHai:%d", releasedHai)
	log.Println(playerInTurn.Hands)
	playerInTurn.DrawnHai = -1
	return releasedHai
}

func (m *MahjongPlayManager) CheckPinfuAndSetRon(releasedHai int) {
	for i, p := range m.playerInfos {
		if i != m.playerIdInTurn {
			p.CanRon = m.PinfuQuery(p.Hands, releasedHai, 27, 27)
			log.Printf("canRon:%d", m.playerIdInTurn)
		}
	}
}

func (m *MahjongPlayManager) PinfuQuery(hands []int, releasedHai, wind int, selfWind int) bool {
	p := PinfuQuery{}
        p.Parse(hands, releasedHai, wind, selfWind)
        log.Println(p)
	return p.Query()
}

func (m *MahjongPlayManager) PlayerInTurnCanRon() bool {
	return m.playerInfos[m.playerIdInTurn].CanRon
}

func (m *MahjongPlayManager) SetFirstPinfuOrder(playerId int) {
	maxOrder := 0
	for _, p := range m.playerInfos {
		if maxOrder < p.FirstPinfuOrder && p.WinnedPinfu() {
			maxOrder = p.FirstPinfuOrder
		}
	}

	if !m.playerInfos[playerId].WinnedPinfu() {
		m.playerInfos[playerId].FirstPinfuOrder = maxOrder + 1;
	}
}

func (m *MahjongPlayManager) CalculateResult() []*Result {
	points := make([]struct {
		playerId int
		point int
	}, playerNumber)

	r := make([]*Result, playerNumber)
	if !m.isEvenGame() {
		for i, p := range m.playerInfos {
			points[i].playerId = i
			points[i].point = p.Point + (10 - p.FirstPinfuOrder) * 10 - p.PlayerId
			log.Println(points[i].point)
		}
		sort.SliceStable(points, func(i, j int) bool {
			return points[i].point > points[j].point
		})
		for i := range m.playerInfos {
			points[i].point = int(math.Floor(float64((points[i].point + 400)/1000))) - 30
		}
		for i := 0; i < playerNumber; i++ {
			points[0].point -= points[i].point
		}

		for i, p := range points {
			r[p.playerId] = &Result{p.point + umaByOrder[i], i + 1}
		}
	} else {
		for i := range points {
			r[i] = &Result{int(math.Floor(float64((points[i].point + 400)/1000))), 0}
		}
	}
	return r
}

func (m *MahjongPlayManager) SendMessageStart() {
	m.SendMessagePlay("start")
}

func (m *MahjongPlayManager) SendMessagePlay(messageType string) {
	for i := range m.sendMessages {
		playerIds := m.GenerateEachPlayerIds(i)
		winds := m.GenerateEachWinds(i)
		points := m.GeneratePoints()
		m.sendMessages[i] = &SendMessage{messageType, &PlayInfo{m.round, m.playerInfos[i], playerIds, winds, points}}
	}
}

func (m *MahjongPlayManager) SendMessageRelease(playerIdInTurnBefore int) {
	log.Printf("send message release playerId:%d", playerIdInTurnBefore)
	for _, v := range m.playerInfos {
		log.Println(v)
	}
	m.sendMessages[playerIdInTurnBefore] = &SendMessage{"release", &m.playerInfos[playerIdInTurnBefore]}
}

func (m *MahjongPlayManager) SendMessageDrawn(releasedHai int) {
	m.playerInfos[m.playerIdInTurn].ReleasedHaiUp = releasedHai
	m.sendMessages[m.playerIdInTurn] = &SendMessage{"drawn", &m.playerInfos[m.playerIdInTurn]}
}

func (m *MahjongPlayManager) SendMessageReleaseOther(playerIdInTurnBefore int, releasedHai int) {
	log.Printf("releasedHai:%d", releasedHai)
	playerIds := m.GenerateEachPlayerIds(playerIdInTurnBefore)
	for i := range m.sendMessages {
		if i != playerIdInTurnBefore {
			r := &ReleasedHaiInfo{playerIds[i], releasedHai, m.playerInfos[i].CanRon}
			m.sendMessages[i] = &SendMessage{"releaseOther", r}
		}
	}
}

func (m *MahjongPlayManager) SendMessageRon() {
	r := make([]*RonInfo, playerNumber)
	r[0] = &RonInfo{26000, 1000}
	r[1] = &RonInfo{25000, 0}
	r[2] = &RonInfo{24000, -1000}
	r[3] = &RonInfo{25000, 0}
	for i := range m.sendMessages {
		m.sendMessages[i] = &SendMessage{"ron", r}
	}
}

func (m *MahjongPlayManager) SendMessageNext() {
	log.Println("SendMessageNext")
	m.SendMessagePlay("next")
}

func (m *MahjongPlayManager) SendMessageResult(r []*Result) {
	for i := range m.sendMessages {
		m.sendMessages[i] = &SendMessage{"result", r}
	}
}

func (m *MahjongPlayManager) WaitNextMessage() {
	m.waitingNext = true
}

func (m *MahjongPlayManager) TriggerNextMessage(f func()) bool {
	m.waitingNextMux.Lock()
	defer m.waitingNextMux.Unlock()
	if m.waitingNext {
		f()
		m.waitingNext = false
		return true
	}
	return false
}

func (m *MahjongPlayManager) RotatePlayer() int {
	playerIdInTurnBefore := m.playerIdInTurn
	m.playerIdInTurn = (m.playerIdInTurn + 1) % playerNumber
	log.Printf("playerId in turn:%d", m.playerIdInTurn);
	return playerIdInTurnBefore
}

func (m *MahjongPlayManager) RotatePlayerWind() {
	for _, p := range m.playerInfos {
		p.Wind = p.Wind.NextPlayerWind()
	}
}

func (m *MahjongPlayManager) RotateRound() {
	if !m.round.isFinalRound() {
		m.round.Round++
	} else if !m.round.isFinalWind() {
		m.round.Wind = m.round.Wind.Next()
		m.round.Round = 1
	}
}

func (m *MahjongPlayManager) continueGame() bool {
	return !m.round.isFinalRound() || (m.isEvenGame() && !m.round.isFinalWind())
}

func (m *MahjongPlayManager) isEvenGame() bool {
	return m.playerInfos[0].Point == m.playerInfos[1].Point && m.playerInfos[0].Point == m.playerInfos[2].Point && m.playerInfos[0].Point == m.playerInfos[3].Point
}

func (p *PlayerInfo) ToBytes() []byte {
	bytes, _ := json.Marshal(p)
	return bytes
}

func (p *PlayerInfo) WinnedPinfu() bool {
	return p.FirstPinfuOrder != playerNumber + 1
}

func (w Wind) Next() Wind {
	return WindList()[(int(w)) % playerNumber]
}

func (w Wind) NextPlayerWind() Wind {
	return WindList()[(int(w) + playerNumber - 2) % playerNumber]
}

func (r *Round) isFinalRound() bool {
	return r.Round == roundNumber
}

func (r *Round) isFinalWind() bool {
	return r.Wind == SOUTH
}

func (s *SendMessage) ToBytes() []byte {
	bytes, _ := json.Marshal(s)
	return bytes
}
/*
func main() {
    m := MahjongPlayManager{}
    m.Init()
}
*/