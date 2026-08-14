package handler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

const (
	ratingKeyCurrent = "recent_rating"
	ratingKeyNew     = "recent_rating_new"
	ratingKeyNext    = "recent_rating_next"
	ratingKeyNextNew = "recent_rating_next_new"
	favoriteMusicKey = "favorite_music"
)

// UpsertUserAll persists every collection carried by UpsertUserAllApi in one transaction.
// The game supplied user ID is authoritative; client supplied GORM primary keys are ignored.
func UpsertUserAll(req model.UpsertUserAllRequest) error {
	if req.UserID == 0 {
		return errors.New("upsert request has no userId")
	}
	if req.UserID&281474976710657 == 281474976710657 {
		return nil
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if len(req.UpsertUserAll.UserData) > 0 {
			user := req.UpsertUserAll.UserData[0]
			user.ID = 0
			user.UserID = req.UserID
			if err := saveDetail(tx, &user); err != nil {
				return err
			}
		} else {
			// SDGA 1.60 can send a partial logout UpsertUserAll with no
			// userData. It still carries scores that must not be discarded. Keep
			// an existing profile intact, or create the minimum profile needed to
			// associate the pending playlogs with this Aime external ID.
			var existing model.UserDetail
			err := tx.Where("user_id = ?", req.UserID).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&model.UserDetail{UserID: req.UserID, IsNetMember: 1}).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}

		if len(req.UpsertUserAll.UserOption) > 0 {
			value := req.UpsertUserAll.UserOption[0]
			value.ID, value.UserID = 0, req.UserID
			if err := saveByUserID(tx, &value); err != nil {
				return err
			}
		}
		if len(req.UpsertUserAll.UserExtend) > 0 {
			value := req.UpsertUserAll.UserExtend[0]
			value.ID, value.UserID = 0, req.UserID
			if err := saveByUserID(tx, &value); err != nil {
				return err
			}
		}

		for _, value := range latestCharacters(req.UpsertUserAll.UserCharacterList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveCharacter(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestMaps(req.UpsertUserAll.UserMapList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveMap(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestLoginBonuses(req.UpsertUserAll.UserLoginBonusList) {
			value.ID, value.UserID, value.IsCurrent = 0, req.UserID, false
			if err := saveLoginBonus(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestItems(req.UpsertUserAll.UserItemList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveItem(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestMusicDetails(req.UpsertUserAll.UserMusicDetailList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveMusicDetail(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestCourses(req.UpsertUserAll.UserCourseList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveCourse(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestCharges(req.UpsertUserAll.UserChargeList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveCharge(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestFriendSeasonRankings(req.UpsertUserAll.UserFriendSeasonRankingList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveFriendSeasonRanking(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestFavorites(req.UpsertUserAll.UserFavoriteList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveFavorite(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestKaleidx(req.UpsertUserAll.UserKaleidxScopeList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveKaleidx(tx, &value); err != nil {
				return err
			}
		}
		for _, value := range latestIntimates(req.UpsertUserAll.UserIntimateList) {
			value.ID, value.UserID = 0, req.UserID
			if err := saveIntimate(tx, &value); err != nil {
				return err
			}
		}

		if len(req.UpsertUserAll.UserRatingList) > 0 {
			rating := req.UpsertUserAll.UserRatingList[0]
			rating.Udemae.ID, rating.Udemae.UserID = 0, req.UserID
			if err := saveByUserID(tx, &rating.Udemae); err != nil {
				return err
			}
			for key, values := range map[string][]model.UserRate{
				ratingKeyCurrent: rating.RatingList, ratingKeyNew: rating.NewRatingList,
				ratingKeyNext: rating.NextRatingList, ratingKeyNextNew: rating.NextNewRatingList,
			} {
				if err := saveGeneralData(tx, req.UserID, key, encodeRating(values)); err != nil {
					return err
				}
			}
		}

		if req.UpsertUserAll.IsNewFavoriteMusicList == "0" {
			musicIDs := make([]string, 0, len(req.UpsertUserAll.UserFavoriteMusicList))
			for _, favorite := range req.UpsertUserAll.UserFavoriteMusicList {
				musicIDs = append(musicIDs, strconv.Itoa(favorite.ID))
			}
			if err := saveGeneralData(tx, req.UserID, favoriteMusicKey, strings.Join(musicIDs, ",")); err != nil {
				return err
			}
		}

		for _, activityGroup := range req.UpsertUserAll.UserActivityList {
			for _, activity := range append(activityGroup.MusicList, activityGroup.PlayList...) {
				if activity.Kind == 0 || activity.ActivityID == 0 {
					continue
				}
				activity.ID, activity.UserID = 0, req.UserID
				if err := tx.Create(&activity).Error; err != nil {
					return err
				}
			}
		}

		playlogs := append(drainPlaylogs(req.UserID), req.UpsertUserAll.UserPlaylogList...)
		playlogs = append(playlogs, req.UserPlaylogList...)
		for _, playlog := range playlogs {
			if err := SaveUserPlaylog(tx, req.UserID, playlog); err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveUserPlaylog applies AquaDX-style de-duplication by game user, music ID and play timestamp.
func SaveUserPlaylog(tx *gorm.DB, userID int64, playlog model.UserPlaylog) error {
	playlog.ID, playlog.UserID = 0, userID
	if playlog.UserPlayDate == "" {
		playlog.UserPlayDate = playlog.PlayDate
	}
	if playlog.CreateDate == "" {
		playlog.CreateDate = playlog.PlayDate
	}
	if playlog.UserPlayDate != "" {
		var existing model.UserPlaylog
		err := tx.Where("user_id = ? AND music_id = ? AND user_play_date = ?", userID, playlog.MusicID, playlog.UserPlayDate).First(&existing).Error
		if err == nil {
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return tx.Create(&playlog).Error
}

func saveDetail(tx *gorm.DB, value *model.UserDetail) error {
	var existing model.UserDetail
	if err := tx.Where("user_id = ?", value.UserID).First(&existing).Error; err == nil {
		value.ID = existing.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}

func saveByUserID(tx *gorm.DB, value interface{}) error {
	switch typed := value.(type) {
	case *model.UserOption:
		var existing model.UserOption
		if err := tx.Where("user_id = ?", typed.UserID).First(&existing).Error; err == nil {
			typed.ID = existing.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	case *model.UserExtend:
		var existing model.UserExtend
		if err := tx.Where("user_id = ?", typed.UserID).First(&existing).Error; err == nil {
			typed.ID = existing.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	case *model.UserUdemae:
		var existing model.UserUdemae
		if err := tx.Where("user_id = ?", typed.UserID).First(&existing).Error; err == nil {
			typed.ID = existing.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return tx.Save(value).Error
}

func saveCharacter(tx *gorm.DB, value *model.UserCharacter) error {
	var x model.UserCharacter
	if err := tx.Where("user_id = ? AND character_id = ?", value.UserID, value.CharacterID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveMap(tx *gorm.DB, value *model.UserMap) error {
	var x model.UserMap
	if err := tx.Where("user_id = ? AND map_id = ?", value.UserID, value.MapID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveLoginBonus(tx *gorm.DB, value *model.UserLoginBonus) error {
	var x model.UserLoginBonus
	if err := tx.Where("user_id = ? AND bonus_id = ?", value.UserID, value.BonusID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveItem(tx *gorm.DB, value *model.UserItem) error {
	var x model.UserItem
	if err := tx.Where("user_id = ? AND item_kind = ? AND item_id = ?", value.UserID, value.ItemKind, value.ItemID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveMusicDetail(tx *gorm.DB, value *model.UserMusicDetail) error {
	var x model.UserMusicDetail
	if err := tx.Where("user_id = ? AND music_id = ? AND level = ?", value.UserID, value.MusicID, value.Level).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveCourse(tx *gorm.DB, value *model.UserCourse) error {
	var x model.UserCourse
	if err := tx.Where("user_id = ? AND course_id = ?", value.UserID, value.CourseID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveCharge(tx *gorm.DB, value *model.UserCharge) error {
	var x model.UserCharge
	if err := tx.Where("user_id = ? AND charge_id = ?", value.UserID, value.ChargeID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}

func saveFriendSeasonRanking(tx *gorm.DB, value *model.UserFriendSeasonRanking) error {
	var x model.UserFriendSeasonRanking
	if err := tx.Where("user_id = ? AND season_id = ?", value.UserID, value.SeasonID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}

func saveFavorite(tx *gorm.DB, value *model.UserFavorite) error {
	var x model.UserFavorite
	if err := tx.Where("user_id = ? AND item_kind = ?", value.UserID, value.ItemKind).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveKaleidx(tx *gorm.DB, value *model.UserKaleidx) error {
	var x model.UserKaleidx
	if err := tx.Where("user_id = ? AND gate_id = ?", value.UserID, value.GateID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveIntimate(tx *gorm.DB, value *model.UserIntimate) error {
	var x model.UserIntimate
	if err := tx.Where("user_id = ? AND partner_id = ?", value.UserID, value.PartnerID).First(&x).Error; err == nil {
		value.ID = x.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Save(value).Error
}
func saveGeneralData(tx *gorm.DB, userID int64, key, value string) error {
	var x model.UserGeneralData
	if err := tx.Where("user_id = ? AND property_key = ?", userID, key).First(&x).Error; err == nil {
		x.PropertyValue = value
		return tx.Save(&x).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return tx.Create(&model.UserGeneralData{UserID: userID, PropertyKey: key, PropertyValue: value}).Error
}

func encodeRating(values []model.UserRate) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%d:%d:%d:%d", v.MusicID, v.Level, v.RomVersion, v.Achievement))
	}
	return strings.Join(parts, ",")
}

func latestCharacters(values []model.UserCharacter) []model.UserCharacter {
	m := map[int]model.UserCharacter{}
	for _, v := range values {
		m[v.CharacterID] = v
	}
	out := make([]model.UserCharacter, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestMaps(values []model.UserMap) []model.UserMap {
	m := map[int]model.UserMap{}
	for _, v := range values {
		m[v.MapID] = v
	}
	out := make([]model.UserMap, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestLoginBonuses(values []model.UserLoginBonus) []model.UserLoginBonus {
	m := map[int]model.UserLoginBonus{}
	for _, v := range values {
		m[v.BonusID] = v
	}
	out := make([]model.UserLoginBonus, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestItems(values []model.UserItem) []model.UserItem {
	m := map[[2]int]model.UserItem{}
	for _, v := range values {
		m[[2]int{v.ItemKind, v.ItemID}] = v
	}
	out := make([]model.UserItem, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestMusicDetails(values []model.UserMusicDetail) []model.UserMusicDetail {
	m := map[[2]int]model.UserMusicDetail{}
	for _, v := range values {
		m[[2]int{v.MusicID, v.Level}] = v
	}
	out := make([]model.UserMusicDetail, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestCharges(values []model.UserCharge) []model.UserCharge {
	m := map[int]model.UserCharge{}
	for _, value := range values {
		m[value.ChargeID] = value
	}
	out := make([]model.UserCharge, 0, len(m))
	for _, value := range m {
		out = append(out, value)
	}
	return out
}

func latestFriendSeasonRankings(values []model.UserFriendSeasonRanking) []model.UserFriendSeasonRanking {
	m := map[int]model.UserFriendSeasonRanking{}
	for _, value := range values {
		m[value.SeasonID] = value
	}
	out := make([]model.UserFriendSeasonRanking, 0, len(m))
	for _, value := range m {
		out = append(out, value)
	}
	return out
}

func latestCourses(values []model.UserCourse) []model.UserCourse {
	m := map[int]model.UserCourse{}
	for _, v := range values {
		m[v.CourseID] = v
	}
	out := make([]model.UserCourse, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestFavorites(values []model.UserFavorite) []model.UserFavorite {
	m := map[int]model.UserFavorite{}
	for _, v := range values {
		m[v.ItemKind] = v
	}
	out := make([]model.UserFavorite, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestKaleidx(values []model.UserKaleidx) []model.UserKaleidx {
	m := map[int]model.UserKaleidx{}
	for _, v := range values {
		m[v.GateID] = v
	}
	out := make([]model.UserKaleidx, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
func latestIntimates(values []model.UserIntimate) []model.UserIntimate {
	m := map[int]model.UserIntimate{}
	for _, v := range values {
		m[v.PartnerID] = v
	}
	out := make([]model.UserIntimate, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
