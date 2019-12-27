from mahjong.hand_calculating.hand import HandCalculator
from mahjong.tile import TilesConverter
from mahjong.hand_calculating.hand_config import HandConfig
from mahjong.meld import Meld
from mahjong.constants import EAST

def check_pinfu(man, pin, sou, honors, player_wind, round_wind, win_tile_type, win_tile_value):
  calculator = HandCalculator()

  tiles = TilesConverter.string_to_136_array(man=man, pin=pin, sou=sou, honors=honors)
  print(tiles)
  win_tile = TilesConverter.string_to_136_array(**{win_tile_type: win_tile_value})[0]

  result = calculator.estimate_hand_value(tiles,
    win_tile,
    config=HandConfig(player_wind=player_wind, round_wind=round_wind))

  if result.yaku is not None:
    for yaku in result.yaku:
      if yaku.name == "Pinfu":
        return True

  return False
