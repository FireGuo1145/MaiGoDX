package ongeki

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
)

// loadGameData reads AquaDX-compatible Ongeki static data without bundling
// the separately licensed game-data files into MaiGoDX. Invalid or missing
// files deliberately behave as empty lists so cabinet boot remains available.
func loadGameData(name string) []map[string]any {
	directory := config("ongeki_game_data_dir", "")
	if directory == "" {
		directory = os.Getenv("ONGEKI_GAME_DATA_DIR")
	}
	if directory == "" {
		return nil
	}
	payload, err := os.ReadFile(filepath.Join(filepath.Clean(directory), filepath.Base(name)))
	if err != nil {
		return nil
	}
	values := []map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if decoder.Decode(&values) != nil {
		return nil
	}
	return values
}

func staticGameListPayload(fileName, key string) map[string]any {
	values := loadGameData(fileName)
	return map[string]any{"length": len(values), key: values}
}

func gameGachas() map[string]any {
	values := loadGameData("game_gacha.json")
	return map[string]any{"length": len(values), "gameGachaList": values, "registIdList": []any{}}
}

func gameGachaCards(gachaID int64) map[string]any {
	values := filterGachaCards(gachaID)
	return map[string]any{"gachaId": gachaID, "length": len(values), "isPickup": false, "gameGachaCardList": values, "emissionList": []any{}, "afterCalcList": []any{}}
}

func filterGachaCards(gachaID int64) []map[string]any {
	all := loadGameData("game_gacha_card.json")
	values := make([]map[string]any, 0)
	for _, value := range all {
		if int64Value(value["gachaId"]) == gachaID {
			values = append(values, value)
		}
	}
	return values
}

func rollGacha(request map[string]any) map[string]any {
	gachaID := int64Value(request["gachaId"])
	times := intValue(request["times"])
	if times <= 0 || times > 100 {
		times = 1
	}
	cards := filterGachaCards(gachaID)
	if len(cards) == 0 {
		return map[string]any{"length": 0, "gameGachaCardList": []any{}}
	}
	byRarity := map[int][]map[string]any{}
	for _, card := range cards {
		byRarity[intValue(card["rarity"])] = append(byRarity[intValue(card["rarity"])], card)
	}
	results := make([]map[string]any, 0, times)
	for index := 0; index < times; index++ {
		rarity := ongekiGachaRarity(rand.Intn(100) + 1)
		pool := byRarity[rarity]
		if len(pool) == 0 {
			pool = cards
		}
		results = append(results, pool[rand.Intn(len(pool))])
	}
	return map[string]any{"length": len(results), "gameGachaCardList": results}
}

func ongekiGachaRarity(value int) int {
	if value <= 76 {
		return 2
	}
	if value <= 97 {
		return 3
	}
	return 4
}
