# 使い方

## Dockerイメージビルド

```
cd /path/to/pinfu-challenge/docker-images
docker image build mahjong --tag mahjong
docker image build mahjong-play-manager --tag mahjong-play-manager
```

## コンテナ起動

```
docker run --rm --name mahjong -v /path/to/pinfu-challenge/hands-calculation:/var/www -p 8000:8000 mahjong /bin/sh -c "python /var/www/api-server.py"
docker run --rm --name mahjong-play-manager -v /path/to/pinfu-challenge/var/www/ -p 8080:8080 mahjong-play-manager /bin/sh -c "cd /var/www/mahjong-play-manager; go run *.go"
```

## mahjong API実行

```
curl -H 'Content-Type:application/json' -d '{"man":"22223333444488","pin":"","sou":"","honors":"","player_wind":27,"round_wind":27,"win_tile_type":"man","win_tile_value":"4"}' http\://localhost:8000
```

## コンテナ停止
```
docker stop mahjong
docker stop mahjong-play-manager
```

# 画像

こちらのサイトで配布されている画材(2が付いていない画材)をmahjong-ui/images以下に配置すると正常に表示されます。
```
https://mj-king.net/sozai/
```

# ルール

## ゲーム全般

```
四人一組で東風戦を一回戦として、東場終了時に得点差があればゲーム終了、全員同点なら南場に入り南四局でゲーム終了
持ち点は各自25000点
座席は接続した順番によって決め、最初に接続した人が起家で、以降順に南家、西家、北家
ワンパイはなく、全ての牌をツモれる※1
ポン、チー、カンなし
九種么九倒牌、四風子連打、流し満貫なし
```

## リーチ

```
なし
```

## フリテン

```
なし(和了牌を河に捨てていてもロン和了可)※2
```

## 連局と流局

```
親の和了時は連チャン
流局時は親は下家に移動(親がテンパイしている場合も)
他家が和了した場合親は下家に移動する(東四局の場合全員同点なら南場に入り、南四局の場合ゲーム終了)
連チャンで積み場が加算され、一本場につき三百点が和了点に加算される
```

## テンパイ

```
流局時の罰符なし※3
```

## ドラ

```
なし
```

## 和了

```
和了は一局一人で、最初にロンした人の和了とする
ツモ和了なし※4
```

## 和了役

```
平和
```

## 順位点

```
25000点持ちの30000点返し
1000点未満は五捨六入
同点時は早い局で和了した者が上位、和了がない場合は起家を基準に起家、下家、対面、上家の順で上位とする※5
ウマはワンツー
終局時四人とも25000点の場合引き分け
```

```
※1 和了しやすくするために全ての牌をツモれるようになっています
※2 ツモ和了がないのでフリテンの形で和了できるようになっています
※3 流局時の罰符が和了より高得点になるので罰符を無くしてあります
※4 門前清自摸和の役がなくツモ和了が20符1飜になるためツモ和了を無くしてあります
※5 早い局での和了放棄を無くすために早い局で和了した人の順位を上げるようにしてあります
```
