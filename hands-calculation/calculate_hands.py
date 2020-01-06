from mahjong.hand_calculating.hand import HandCalculator
from mahjong.tile import TilesConverter
from mahjong.hand_calculating.hand_config import HandConfig

import json

def check_pinfu(man, pin, sou, honors, player_wind, round_wind, win_tile_type, win_tile_value):
  calculator = HandCalculator()

  tiles = TilesConverter.string_to_136_array(man=man, pin=pin, sou=sou, honors=honors)
  print(tiles)
  win_tile = TilesConverter.string_to_136_array(**{win_tile_type: win_tile_value})[0]

  config = HandConfig(player_wind=player_wind, round_wind=round_wind)
  result = calculator.estimate_hand_value(tiles, win_tile, config=config)

  if result.yaku is not None:
    for yaku in result.yaku:
      if yaku.name == "Pinfu":
        cost = 1500 if config.is_dealer else 1000
        return [json.dumps({'isPinfu':True,'cost':cost}).encode("utf-8")]

  return [json.dumps({'isPinfu':False,'cost':0}).encode("utf-8")]
