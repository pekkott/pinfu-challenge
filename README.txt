docker image build mahjong --tag mahjong
docker image build mahjong-play-manager --tag mahjong-play-manager

docker run -p 8000:8000 -v /Volumes/HD-LCU3/workspace/pinfu-challenge/hands-calculation:/var/www mahjong /bin/sh -c "python /var/www/api-server.py"

docker run -p 8000:8000 -v /Volumes/HD-LCU3/workspace/pinfu-challenge/hands-calculation:/var/www mahjong /bin/sh -c "python /var/www/api-server.py"

docker run --rm --name mahjong-play-manager -v /Volumes/HD-LCU3/workspace/pinfu-challenge/var/www/ -p 8080:8080 mahjong-play-manager /bin/sh -c "cd /var/www/mahjong-play-manager; go run *.go"

docker stop mahjong-play-manager

curl -H 'Content-Type:application/json' -d '{"man":"22223333444488","pin":"","sou":"","honors":"","player_wind":27,"round_wind":27,"win_tile_type":"man","win_tile_value":"4"}' http://localhost:8000
