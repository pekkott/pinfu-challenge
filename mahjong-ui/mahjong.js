class Mahjong {
    static get GET_WIND() {
        return [
            'east',
            'south',
            'west',
            'north'
        ];
    }

    static get GET_POSITION() {
        return [
            'self',
            'down',
            'opposite',
            'up'
        ];
    }

    static get POSITION_SELF() {
        return Mahjong.GET_POSITION[0];
    }

    static get POSITION_DOWN() {
        return Mahjong.GET_POSITION[1];
    }

    static get POSITION_OPPOSITE() {
        return Mahjong.GET_POSITION[2];
    }

    static get POSITION_UP() {
        return Mahjong.GET_POSITION[3];
    }

    static get HO_ROW_SIZE() {
        return 6;
    }

    static get HO_COLUMN_SIZE() {
        return 4;
    }

    static get HO_SIZE() {
        return this.HO_ROW_SIZE * this.HO_COLUMN_SIZE;
    }

    static get SELF() {
        return 0;
    }

    static get OPPOSITE() {
        return 1;
    }

    static get DOWN() {
        return 2;
    }

    static get UP() {
        return 3;
    }

    static get HAI_TYPES() {
        return [
            'ms1', 'ms2', 'ms3', 'ms4', 'ms5', 'ms6', 'ms7', 'ms8', 'ms9',
            'ps1', 'ps2', 'ps3', 'ps4', 'ps5', 'ps6', 'ps7', 'ps8', 'ps9',
            'ss1', 'ss2', 'ss3', 'ss4', 'ss5', 'ss6', 'ss7', 'ss8', 'ss9',
            'ji_e', 'ji_s', 'ji_w', 'ji_n', 'no', 'ji_h', 'ji_c'
        ];
    }

    static get MODAL_DURATION() {
        return 2000;
    }
}

class MahjongManager {
    constructor() {
        this.wind = 0;
        this.players = [
            new PlayerSelf(0),
            new PlayerDown(1),
            new PlayerOpposite(2),
            new PlayerUp(3)
        ];
        this.ron = new Ron('#modal-ron');
        this.result = new Result('#modal-result');
        this.webSocketManager = new WebSocketManager(this);

        $('#hands-hai-self').on('click', (event) => this.release(event));
        $('#hai-drawn-self').on('click', (event) => this.releaseDrawnHai(event));
        $('#fu-self').on('click', (event) => this.releaseDrawnHai(event));
        $('#close-modal-ron').on('click', (event) => this.closeModalRon());
        $('#close-modal-result').on('click', (event) => this.closeModalResult());
    }

    setPlayersId(playerIds) {
        this.players.forEach(function(player, i) {
            player.playerId = playerIds[i];
        });
    }

    setRound(round) {
        var windTable = {
            1:"東",
            2:"南"
        };
        var roundTable = {
            1:"一",
            2:"二",
            3:"三",
            4:"四"
        };
        $('#round').each(function(item) {
            item.innerHTML = windTable[round.wind] + roundTable[round.round] + "局";
        });
        if (round.subRound > 0) {
            $('#sub-round').each(function(item) {
                item.classList.remove("display-none");
            });
            $('.sub-round').each(function(item) {
                item.innerHTML = round.subRound;
            });
        }
    }

    setWinds(winds) {
        this.players.forEach(function(player, i) {
            player.wind.wind = winds[i];
        });
    }

    updatePlayerPoints(ronInfo) {
        this.ron.updatePlayerPoints(ronInfo);
    }

    updatePoints(points) {
        this.players.forEach(function(player, i) {
            player.point.point = points[i];
        });
    }

    updatePointsByRonInfo(ronInfo) {
        var points = this.ron.pointsByRonInfo(ronInfo);
        this.updatePoints(points);
    }

    canRelease(event) {
        return this.players[0].canRelease();
    }

    release(event) {
        self = this;
        this.players.forEach(function(player) {
            if (player.canRelease()) {
                console.log("release:" + event.target.value);
                var released = player.release(event.target.value);
                player.showHo();
                player.updateHands(self);
                player.showHands();
                player.disableRelease();
            }
        });
    }

    releaseDrawnHai(event) {
        this.players.forEach(function(player) {
            if (player.canRelease()) {
                player.releaseDrawnHai();
                player.showHo();
            }
        });
    }

    updatePlayersPoint() {
    }

    updatePlayerHands(info) {
        this.players[0].hands = info.hands;
        this.players[0].drawnHai = info.drawnHai;
    }

    showHands() {
        this.players.forEach(function(player) {
            player.showHands();
        });
    }

    showDrawnHai() {
        this.players[0].showDrawnHai();
    }

    showWind() {
        this.players.forEach(function(player) {
            player.wind.show();
        });
    }

    showPoint() {
        this.players.forEach(function(player) {
            player.point.show();
        });
    }

    showRon() {
        this.ron.showModal();
        this.ron.setModalTimeout(this.webSocketManager);
    }

    showResult(result) {
        this.result.update(result);
        this.result.showModal();
    }

    hideDrawnHai() {
        this.players[0].hideDrawnHai();
    }

    closeModalRon() {
        this.ron.showModal();
    }

    closeModalResult() {
        this.result.showModal();
    }

    clearHo() {
        this.players.forEach(function(player) {
            player.clearHo();
            player.showHo();
        });
    }
}

class Player {
    constructor(position, positionClassName, wind, pointClassName) {
        this.playerId = -1;
        this.position = position;
        this.positionClassName = positionClassName;
        this.wind = new Wind(wind, positionClassName);
        this.point = new Point(positionClassName);
        this.haisReleased = new Map();
    }

    updateHands() {
    }

    showHands() {
    }

    showWind() {
        this.wind.show();
    }

    showHo() {
        var self = this;
        $('.hai-' + self.positionClassName).each(function(item, i) {
            self.showHai(item, self.haisReleased.get(i));
            self.modifyHaiLayout(item, i);
        });
    }

    toOrderedIndex() {
        return this.haisReleased.size;
    }

    showHai(item, haiId) {
        var imageName = Mahjong.HAI_TYPES[this.toHaiType(haiId)];
        if (imageName) {
            item.style.setProperty('--url-hai', "url('/mahjong-ui/images/p_" + imageName + '_' + (this.position + 1) + ".gif')");
        }
    }

    toHaiType(i) {
        return Math.floor(i/4);
    }

    modifyHaiLayout(i) {
    }

    canRelease() {
        return false;
    }

    release(haiIndex) {
    }

    releaseOther(releaseHai) {
        if (releaseHai != -1) {
            var i = this.toOrderedIndex();
            this.haisReleased.set(i, releaseHai);
        }
    }

    releaseDrawnHai() {
    }

    clearHo() {
        this.haisReleased = new Map();
        $('.hai-' + this.positionClassName).each(function(item, i) {
            item.style.setProperty('--url-hai', "");
        });
    }
}

class PlayerOpposite extends Player {
    constructor(wind) {
        super(Mahjong.OPPOSITE, Mahjong.POSITION_OPPOSITE, wind);
    }

    toOrderedIndex() {
        var order = this.haisReleased.size;

        return Mahjong.HO_SIZE - 1 - order;
    }
}

class PlayerUp extends Player {
    constructor(wind) {
        super(Mahjong.UP, Mahjong.POSITION_UP, wind);
    }

    toOrderedIndex() {
        var order = this.haisReleased.size;

        return (order*4 + 3) % Mahjong.HO_SIZE - Math.floor(order/6);
    }

    modifyHaiLayout(item, i) {
        var haiTypeClass = this.chooseHaiTypeClass(i);
        if (haiTypeClass) {
            item.className = 'hai-' + this.positionClassName + ' hai-horizontal-' + haiTypeClass;
        }
    }

    chooseHaiTypeClass(i) {
        if (this.haisReleased.has(i) && this.isAtBottom(i)) {
            return 'bottom';
        } else if (!this.haisReleased.has(i) && this.hasUpperHai(i)) {
            return 'bottom-parts';
        } else {
            return 'middle';
        }

        return null;
    }

    hasUpperHai(i) {
        return this.haisReleased.has(i - Mahjong.HO_COLUMN_SIZE);
    }

    isAtBottom(i) {
        return i > Mahjong.HO_SIZE - Mahjong.HO_COLUMN_SIZE - 1;
    }
}

class PlayerDown extends Player {
    constructor(wind) {
        super(Mahjong.DOWN, Mahjong.POSITION_DOWN, wind);
    }

    toOrderedIndex() {
        var order = this.haisReleased.size;

        return (Mahjong.HO_SIZE - 1 - ((order + 1) % Mahjong.HO_ROW_SIZE)*Mahjong.HO_COLUMN_SIZE + Math.ceil((order + 1)/Mahjong.HO_ROW_SIZE)) % Mahjong.HO_SIZE;
    }
}

class PlayerSelf extends Player {
    constructor(wind) {
        super(Mahjong.SELF, Mahjong.POSITION_SELF, wind);
        this.hands = [];
        this.drawnHai = null;
        this.i = 0;
    }

    updateHands(mahjongManager) {
        this.i++;
        this.hands.sort(function(a, b) {return a - b;});
    }

    showHands() {
        var self = this;
        $('.hai-in-hands').each(function(item, i) {
            self.showHai(item, self.hands[i]);
        });
    }

    showDrawnHai() {
        var self = this;
        $('#hai-drawn-self').each(function(item) {
            console.log("drawnHai:" + self.drawnHai);
            self.showHai(item, self.drawnHai);
        });
    }

    hideDrawnHai(item) {
        $('#hai-drawn-self').each(function(item) {
            item.style.setProperty('--url-hai', "");
        });
    }

    modifyHaiLayout(item, i) {
        var haiTypeClass = this.chooseHaiTypeClass(i);
        if (haiTypeClass) {
            item.className = 'hai-' + this.positionClassName + ' hai-vertical-' + haiTypeClass;
        }
    }

    chooseHaiTypeClass(i) {
        if (!this.hasNextRow(i)) {
            if (this.haisReleased.has(i)) {
                return 'bottom';
            } else if (this.hasSelfRow(i) && this.hasUpperHai(i)) {
                return 'bottom-parts';
            }
        } else if (this.haisReleased.has(i) || !this.hasBelowHai(i)) {
            return 'middle';
        }

        return null;
    }

    hasSelfRow(i) {
        return this.haisReleased.has(i - (i % Mahjong.HO_ROW_SIZE));
    }

    hasNextRow(i) {
        return this.haisReleased.has(i + Mahjong.HO_ROW_SIZE - (i % Mahjong.HO_ROW_SIZE));
    }

    hasUpperHai(i) {
        return this.haisReleased.has(i - Mahjong.HO_ROW_SIZE);
    }

    hasBelowHai(i) {
        return this.haisReleased.has(i + Mahjong.HO_ROW_SIZE);
    }

    release(haiIndex) {
        if (this.canRelease()) {
            var released = this.hands.splice(haiIndex, 1)[0];
            var i = this.toOrderedIndex();
            this.haisReleased.set(i, released);
            this.drawnHai = -1;
        }
    }

    releaseDrawnHai() {
        if (this.canRelease()) {
            console.log("release:" + this.drawnHai);
            var i = this.toOrderedIndex();
            this.haisReleased.set(i, this.drawnHai);
            this.drawnHai = -1;
        }
    }

    canRelease() {
        return this.drawnHai != -1;
    }

    disableRelease() {
        this.drawnHai = -1;
    }

    clearHo() {
        super.clearHo();
        $('.hai-' + this.positionClassName).each(function(item, i) {
            item.className = "hai-self hai-vertical-middle";
        });
    }
}

class Wind {
    constructor(wind, positionClassName) {
        this.wind = wind;
        this.positionClassName = positionClassName;
    }

    show() {
        self = this;
        $('#fu-' + self.positionClassName).each(function(item, i) {
            item.firstElementChild.className = 'fu fu-' + Mahjong.GET_WIND[self.wind - 1] + '-' + self.positionClassName;
        });
    }
}

class Point {
    constructor(positionClassName) {
        this.point = 0;
        this.positionClassName = positionClassName;
    }

    show() {
        self = this;
        $('#point-' + self.positionClassName).each(function(item, i) {
            item.innerHTML = self.point;
        });
    }
}

class Modal {
    constructor(modalId) {
        this.modalId = modalId;
    }

    showModal() {
        $(this.modalId).each(function(item, i) {
            item.classList.toggle("show-modal");
        });
    }

    setModalTimeout(webSocketManager) {
        $(this.modalId).each(function(item, i) {
            setTimeout(function() {
                item.classList.remove("show-modal");
                webSocketManager.sendNext();
            }, Mahjong.MODAL_DURATION);
        });
    }
}

class Ron extends Modal {
    updatePlayerPoints(ronInfo) {
        $('.player-point').each(function(item, i) {
            item.innerHTML = ronInfo[i].point;
        });
        $('.player-point-diff').each(function(item, i) {
            if (ronInfo[i].pointDiff == 0) {
                item.innerHTML = "";
            } else {
                item.innerHTML = (ronInfo[i].pointDiff > 0 ? "+" : "") + ronInfo[i].pointDiff;
            }
        });
    }

    pointsByRonInfo(ronInfo) {
        var points = [];
        ronInfo.forEach(function(ron) {
            points.push(ron.point);
        });

        return points;
    }

    showButton() {
        $('#ron').each(function(item) {
            item.classList.remove("display-none");
        });
    }
}

class Result extends Modal {
    update(result) {
        $('.result-order').each(function(item, i) {
            item.innerHTML = result[i].order > 0 ? result[i].order : "-";
        });
        $('.result-point').each(function(item, i) {
            item.innerHTML = (result[i].point > 0 ? "+" : "") + result[i].point;
        });
    }
}

class WebSocketManager {
    constructor(mahjongManager) {
        var self = this;
        self.mahjongManager = mahjongManager;
        this.messageHandlers = [
            {type: "start", handler: this.receiveStart},
            {type: "release", handler: this.receiveRelease},
            {type: "drawn", handler: this.receiveDrawn},
            {type: "releaseOther", handler: this.receiveReleaseOther},
            {type: "ron", handler: this.receiveRon},
            {type: "next", handler: this.receiveNext},
            {type: "result", handler: this.receiveResult}
        ];
        if (window["WebSocket"]) {
            self.conn = new WebSocket("ws://" + document.location.host + "/ws");
            self.conn.onmessage = function (evt) {
                var message = JSON.parse(evt.data);
                self.messageHandlers.forEach(function(item) {
                    if (item.type == message["type"]) {
                        console.log("received message type:" + message["type"]);
                        item.handler(mahjongManager, message["values"]);
                    }
                });
            }
        }
        $('#fu-self').on('click', (event) => this.debug(event));
        $('#hands-hai-self').on('click', (event) => this.sendRelease(event));
        $('#hai-drawn-self').on('click', (event) => this.sendRelease(event));
        $('#ron').on('click', (event) => this.sendRon(event));
        $('#debug-ron').on('click', (event) => this.debugRon(event));
        $('#debug-start').on('click', (event) => this.debugStart(event));
        $('#debug-next').on('click', (event) => this.debugNext(event));
        $('#debug-result').on('click', (event) => this.debugResult(event));
    }

    receiveStart(mahjongManager, playInfo) {
        console.log(playInfo);
        mahjongManager.setPlayersId(playInfo.playerIds);
        mahjongManager.setRound(playInfo.round);
        mahjongManager.setWinds(playInfo.winds);
        mahjongManager.updatePoints(playInfo.points);
        mahjongManager.updatePlayerHands(playInfo.playerInfo);
        mahjongManager.showPoint();
        mahjongManager.showWind();
        mahjongManager.showHands();
        mahjongManager.showDrawnHai();
    }

    receiveRelease(mahjongManager, playerInfo) {
        console.log(playerInfo);
        mahjongManager.updatePlayerHands(playerInfo);
        mahjongManager.showHands();
        mahjongManager.hideDrawnHai();
    }

    receiveDrawn(mahjongManager, playerInfo) {
        console.log(playerInfo);
        mahjongManager.updatePlayerHands(playerInfo);
        mahjongManager.showDrawnHai();
        mahjongManager.players[3].releaseOther(playerInfo.releasedHaiUp);
        mahjongManager.players[3].showHo();
    }

    receiveReleaseOther(mahjongManager, releaseddHaiInfo) {
        console.log("release other position:" + releaseddHaiInfo.playerPosition);
        console.log("release other hai:" + releaseddHaiInfo.releasedHai);
        if (releaseddHaiInfo.canRon) {
            mahjongManager.ron.showButton();
        }
        mahjongManager.players[releaseddHaiInfo.playerPosition].releaseOther(releaseddHaiInfo.releasedHai);
        mahjongManager.players[releaseddHaiInfo.playerPosition].showHo();
    }

    receiveRon(mahjongManager, ronInfo) {
        console.log(ronInfo);
        mahjongManager.updatePlayerPoints(ronInfo);
        mahjongManager.showRon();
        mahjongManager.updatePointsByRonInfo(ronInfo);
        mahjongManager.showPoint();
    }

    receiveNext(mahjongManager, playInfo) {
        mahjongManager.setRound(playInfo.round);
        mahjongManager.setWinds(playInfo.winds);
        mahjongManager.updatePoints(playInfo.points);
        mahjongManager.updatePlayerHands(playInfo.playerInfo);
        mahjongManager.showPoint();
        mahjongManager.showWind();
        mahjongManager.showHands();
        mahjongManager.showDrawnHai();
        mahjongManager.clearHo();
    }

    receiveResult(mahjongManager, resultInfo) {
        console.log(resultInfo);
        mahjongManager.showResult(resultInfo);
    }

    sendRelease(event) {
        if (this.mahjongManager.canRelease()) {
            console.log("send:" + event.target.value);
            this.conn.send(JSON.stringify({operation: "release", target: event.target.value}));
        }
    }

    sendRon(event) {
        console.log("send ron");
        this.conn.send(JSON.stringify({operation: "ron", target: -1}));
    }

    sendNext() {
        this.conn.send(JSON.stringify({operation: "next", target: -1}));
    }

    debug(event) {
        this.conn.send(JSON.stringify({operation: "release", target: -1}));
    }

    debugRon(event) {
        this.conn.send(JSON.stringify({operation: "ron", target: -1}));
    }

    debugStart(event) {
        this.conn.send(JSON.stringify({operation: "start", target: event.target.value}));
    }

    debugNext(event) {
        this.conn.send(JSON.stringify({operation: "next", target: event.target.value}));
    }

    debugResult(event) {
        this.conn.send(JSON.stringify({operation: "result", target: event.target.value}));
    }
}
