package ongeki

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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

func gamePresents() map[string]any {
	values := loadGameData("game_present.json")
	for _, value := range values {
		value["startDate"] = "2000-01-01 05:00:00.0"
		value["endDate"] = "2099-01-01 05:00:00.0"
	}
	return map[string]any{"length": len(values), "gamePresentList": values}
}

func gameGachas() map[string]any {
	values := loadGameData("game_gacha.json")
	return map[string]any{"length": len(values), "gameGachaList": values, "registIdList": []any{}}
}

func findGacha(gachaID int64) (map[string]any, bool) {
	for _, gacha := range loadGameData("game_gacha.json") {
		if int64Value(gacha["gachaId"]) == gachaID {
			return gacha, true
		}
	}
	return nil, false
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

func rollGacha(userID int64, request map[string]any) map[string]any {
	gachaID := int64Value(request["gachaId"])
	times := intValue(request["times"])
	if times <= 0 || times > 100 {
		times = 1
	}
	cards := gachaCardPool(gachaID)
	if len(cards) == 0 {
		return map[string]any{"length": 0, "gameGachaCardList": []any{}}
	}
	byRarity := map[int][]map[string]any{}
	for _, card := range cards {
		byRarity[intValue(card["rarity"])] = append(byRarity[intValue(card["rarity"])], card)
	}
	rarityResults := make([]int, 0, times)
	if gachaState := firstRecordByKey(userID, "userGachaList", strconv.FormatInt(gachaID, 10)); len(gachaState) > 0 && ((times == 5 && intValue(gachaState["fiveGachaCnt"]) == 0) || times == 11) {
		rarityResults = append(rarityResults, 3)
	}
	for len(rarityResults) < times {
		rarityResults = append(rarityResults, ongekiGachaRarity(rand.Intn(100)+1))
	}
	results := make([]map[string]any, 0, times)
	for _, rarity := range rarityResults {
		pool := byRarity[rarity]
		if len(pool) == 0 {
			pool = cards
		}
		results = append(results, pool[rand.Intn(len(pool))])
	}
	return map[string]any{"length": len(results), "gameGachaCardList": results}
}

func gachaCardPool(gachaID int64) []map[string]any {
	values := filterGachaCards(gachaID)
	if gacha, ok := findGacha(gachaID); ok {
		kind := intValue(gacha["kind"])
		if kind != 0 && kind != 3 && gachaID != 1112 {
			values = append(values, filterGachaCards(1112)...)
		}
	}
	return values
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
