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
	tileInMountNumber = 136
	tileInDistributionClusterNumber = 4
	tileInHandNumber = 13
	tileIdNone = -1
	playerNumber = 4
	playerIdNone = -1
	roundNumber = 4
	pointStart = 25000
	costBySubRound = 300
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
	isDealerWin bool
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
	DrawnTile int `json:"drawnTile"`
	DiscardedTileUp int `json:"discardedTileUp"`
	PinfuInfo *PinfuInfo `json:"-"`
}

type DiscardedTileInfo struct {
	PlayerPosition int `json:"playerPosition"`
	DiscardedTile int `json:"discardedTile"`
	CanRon bool `json:"canRon"`
}

type CanRonInfo struct {
	CanRon bool `json:"canRon"`
}

type RonInfo struct {
	Point int `json:"point"`
	PointDiff int `json:"pointDiff"`
}

type DrawnRoundInfo struct {
	RonInfo []*RonInfo `json:"ronInfo"`
	DiscardedTileInfo *DiscardedTileInfo `json:"discardedTileInfo"`
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
	m.playerNumber = playerIdNone
	m.playerInfos = make([]*PlayerInfo, playerNumber)
	for i := range m.playerInfos {
		m.playerInfos[i] = &PlayerInfo{i, pointStart, playerNumber + 1, WindList()[i], make([]int, tileInHandNumber), tileIdNone, tileIdNone, &PinfuInfo{false, 0}}
	}
	m.InitSeed()
	m.waitingNext = false
	m.isDealerWin = false
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

func (m *MahjongPlayManager) InitRound() {
	m.InitPlayerInfos()
	m.InitPlayerIdInTrun()
	m.InitMount()
	m.InitHands()
	m.DistributeTile()
	m.isDealerWin = false
}

func (m *MahjongPlayManager) InitPlayerInfos() {
	for _, p := range m.playerInfos {
		p.DrawnTile = tileIdNone
		p.DiscardedTileUp = tileIdNone
		p.PinfuInfo = &PinfuInfo{false, 0}
	}
}

func (m *MahjongPlayManager) InitMount() {
	m.mountPosition = 0
	m.mount = make([]int, tileInMountNumber)
	for i := range m.mount {
		m.mount[i] = i
	}
	rand.Shuffle(len(m.mount), func(i, j int) {
		m.mount[i], m.mount[j] = m.mount[j], m.mount[i]
	})
	log.Println(m.mount)
}

func (m *MahjongPlayManager) InitHands() {
	clusterNumber := (tileInHandNumber - 1)/tileInDistributionClusterNumber
	for i := 0; i < clusterNumber; i++ {
		for j := 0; j < playerNumber; j++ {
			targetId := (m.playerIdInTurn + j) % playerNumber
			for k := 0; k < tileInDistributionClusterNumber; k++ {
				m.playerInfos[targetId].Hands[i*tileInDistributionClusterNumber + k] = m.mount[m.mountPosition]
				m.mountPosition++
			}
		}
	}

	for i := 0; i < playerNumber; i++ {
		targetId := (m.playerIdInTurn + i) % playerNumber
		m.playerInfos[targetId].Hands[tileInHandNumber - 1] = m.mount[m.mountPosition]
		m.mountPosition++
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

func (m *MahjongPlayManager) GenerateEachPoints(playerId int) []int {
	points := make([]int, playerNumber)
	for i, p := range m.playerInfos {
		points[(playerNumber - playerId + i) % playerNumber] = p.Point
	}
	return points
}

func (m *MahjongPlayManager) DistributeTile() {
	m.playerInfos[m.playerIdInTurn].DrawnTile = m.mount[m.mountPosition]
	m.mountPosition++
}

func (m *MahjongPlayManager) CanDistributeTile() bool {
	return m.mountPosition < tileInMountNumber
}

func (m *MahjongPlayManager) DiscardTile(position int) int {
	playerInTurn := m.playerInfos[m.playerIdInTurn]
	discardedTile := playerInTurn.DrawnTile
	if position >= 0 && position < len(playerInTurn.Hands) {
		log.Printf("discard position:%d", position)
		discardedTile = playerInTurn.Hands[position]
		playerInTurn.Hands[position] = playerInTurn.DrawnTile
		sort.Ints(playerInTurn.Hands)
	}
	log.Printf("discard drawnTile:%d", playerInTurn.DrawnTile)
	log.Printf("discardedTile:%d", discardedTile)
	log.Println(playerInTurn.Hands)
	playerInTurn.DrawnTile = tileIdNone
	return discardedTile
}

func (m *MahjongPlayManager) CheckPinfuAndSetRon(discardedTile int) bool {
	canRon := false
	for i, p := range m.playerInfos {
		if i != m.playerIdInTurn {
			p.PinfuInfo = m.PinfuQuery(p.Hands, discardedTile, int(m.round.Wind), int(p.Wind))
			if p.PinfuInfo.IsPinfu {
				canRon = true
			}
			log.Printf("canRon:%d", m.playerIdInTurn)
		}
	}
	return canRon
}

func (m *MahjongPlayManager) PinfuQuery(hands []int, discardedTile, wind int, selfWind int) *PinfuInfo {
	p := PinfuQuery{}
	p.Parse(hands, discardedTile, selfWind, wind)
	log.Println(p)
	return p.Query()
}

func (m *MahjongPlayManager) CalculateRonInfo(playerId int) []*RonInfo {
	r := make([]*RonInfo, playerNumber)
	for i, p := range m.playerInfos {
		r[i] = &RonInfo{p.Point, 0}
	}
	cost := m.playerInfos[playerId].PinfuInfo.Cost + costBySubRound*m.round.SubRound
	r[playerId].Update(cost)
	r[m.playerIdInTurn].Update(-cost)
	return r
}

func (m *MahjongPlayManager) UpdatePlayersPoint(r []*RonInfo) {
	for _, p := range m.playerInfos {
		p.Point = r[p.PlayerId].Point
	}
}

func (m *MahjongPlayManager) SetFirstPinfuOrder(playerId int) {
	maxOrder := 0
	for _, p := range m.playerInfos {
		if maxOrder < p.FirstPinfuOrder && p.WinnedPinfu() {
			maxOrder = p.FirstPinfuOrder
		}
	}

	if !m.playerInfos[playerId].WinnedPinfu() {
		m.playerInfos[playerId].FirstPinfuOrder = maxOrder + 1
	}
}

func (m *MahjongPlayManager) CalculateResult() []*Result {
	points := make([]struct {
		playerId int
		point int
	}, playerNumber)

	r := make([]*Result, playerNumber)
	if !m.isDrawnGame() {
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
		points := m.GenerateEachPoints(i)
		m.sendMessages[i] = &SendMessage{messageType, &PlayInfo{m.round, m.playerInfos[i], playerIds, winds, points}}
	}
}

func (m *MahjongPlayManager) SendMessageDiscard(playerIdInTurnBefore int) {
	log.Printf("send message discard playerId:%d", playerIdInTurnBefore)
	for _, v := range m.playerInfos {
		log.Println(v)
	}
	m.sendMessages[playerIdInTurnBefore] = &SendMessage{"discard", &m.playerInfos[playerIdInTurnBefore]}
}

func (m *MahjongPlayManager) SendMessageDrawn(discardedTile int) {
	m.playerInfos[m.playerIdInTurn].DiscardedTileUp = discardedTile
	m.sendMessages[m.playerIdInTurn] = &SendMessage{"drawn", &m.playerInfos[m.playerIdInTurn]}
}

func (m *MahjongPlayManager) SendMessageDiscardOther(playerIdInTurnBefore int, discardedTile int) {
	log.Printf("discardedTile:%d", discardedTile)
	playerIds := m.GenerateEachPlayerIds(playerIdInTurnBefore)
	for i := range m.sendMessages {
		if i != playerIdInTurnBefore {
			r := &DiscardedTileInfo{playerIds[i], discardedTile, m.playerInfos[i].PinfuInfo.IsPinfu}
			m.sendMessages[i] = &SendMessage{"discardOther", r}
		}
	}
}

func (m *MahjongPlayManager) SendMessageCanRon() {
	for i, p := range m.playerInfos {
		if i != m.playerIdInTurn {
			m.sendMessages[i] = &SendMessage{"canRon", CanRonInfo{p.PinfuInfo.IsPinfu}}
		}
	}
}

func (m *MahjongPlayManager) SendMessageRon(r []*RonInfo) {
	for i := range m.sendMessages {
		m.sendMessages[i] = &SendMessage{"ron", r}
	}
}

func (m *MahjongPlayManager) SendMessageSkip() {
	for i := range m.sendMessages {
		if i != m.playerIdInTurn {
			m.sendMessages[i] = &SendMessage{"skip", ""}
		}
	}
}

func (m *MahjongPlayManager) SendMessageDrawnRound(discardedTile int) {
	r := make([]*RonInfo, playerNumber)
	for i, p := range m.playerInfos {
		r[i] = &RonInfo{p.Point, 0}
	}
	for i := range m.sendMessages {
		m.sendMessages[i] = &SendMessage{"drawnRound", &DrawnRoundInfo{r, &DiscardedTileInfo{(playerNumber - m.playerIdInTurn + i) % playerNumber, discardedTile, false}}}
	log.Printf("DiscardedTileInfo:%d", (playerNumber - m.playerIdInTurn + i) % playerNumber)
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

func (m *MahjongPlayManager) IsDealerWin() bool {
	return m.isDealerWin
}

func (m *MahjongPlayManager) DealerWin(playerId int) {
	m.isDealerWin = m.playerInfos[playerId].Wind == EAST
}

func (m *MahjongPlayManager) RotatePlayer() int {
	playerIdInTurnBefore := m.playerIdInTurn
	m.playerIdInTurn = (m.playerIdInTurn + 1) % playerNumber
	log.Printf("playerId in turn:%d", m.playerIdInTurn)
	return playerIdInTurnBefore
}

func (m *MahjongPlayManager) RotatePlayerWind() {
	for _, p := range m.playerInfos {
		p.Wind = p.Wind.NextPlayerWind()
	}
}

func (m *MahjongPlayManager) NextSubRound() {
	m.round.NextSubRound()
}

func (m *MahjongPlayManager) ResetSubRound() {
	m.round.ResetSubRound()
}

func (m *MahjongPlayManager) RotateRound() {
	if !m.round.IsFinalRound() {
		m.round.Round++
	} else if !m.round.IsFinalWind() {
		m.round.Wind = m.round.Wind.Next()
		m.round.Round = 1
	}
}

func (m *MahjongPlayManager) continueGame() bool {
	return !m.round.IsFinalRound() || (m.isDrawnGame() && !m.round.IsFinalWind())
}

func (m *MahjongPlayManager) isDrawnGame() bool {
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

func (r *Round) NextSubRound() {
	r.SubRound++
}

func (r *Round) ResetSubRound() {
	r.SubRound = 0
}

func (r *Round) IsFinalRound() bool {
	return r.Round == roundNumber
}

func (r *Round) IsFinalWind() bool {
	return r.Wind == SOUTH
}

func (r *RonInfo) Update(cost int) {
	r.Point = r.Point + cost
	r.PointDiff = cost
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