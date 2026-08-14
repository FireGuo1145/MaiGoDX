package model

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// UserDetail persists a maimai player profile. UserID is the game card external ID.
type UserDetail struct {
	gorm.Model               `json:"-"`
	UserID                   int64  `gorm:"uniqueIndex;not null" json:"userId"`
	UserName                 string `json:"userName"`
	IsNetMember              int    `json:"isNetMember"`
	EquipGlassesID           int    `json:"equipGlassesId"`
	EquipBackGroundID        int    `json:"equipBackGroundId"`
	EquipNamePlateID         int    `json:"equipNamePlateId"`
	EquipFrameID             int    `json:"equipFrameId"`
	EquipIconID              int    `json:"equipIconId"`
	IconID                   int    `json:"iconId"`
	PlateID                  int    `json:"plateId"`
	TitleID                  int    `json:"titleId"`
	PartnerID                int    `json:"partnerId"`
	FrameID                  int    `json:"frameId"`
	SelectMapID              int    `json:"selectMapId"`
	TotalAwake               int    `json:"totalAwake"`
	GradeRating              int    `json:"gradeRating"`
	MusicRating              int    `json:"musicRating"`
	Rating                   int    `json:"rating"`
	MaxRating                int    `json:"maxRating"`
	GradeRank                int    `json:"gradeRank"`
	ClassRank                int    `json:"classRank"`
	CourseRank               int    `json:"courseRank"`
	ContentBit               int64  `json:"contentBit"`
	PlayCount                int    `json:"playCount"`
	EventWatchedDate         string `json:"eventWatchedDate"`
	LastGameID               string `json:"lastGameId"`
	LastRomVersion           string `json:"lastRomVersion"`
	LastDataVersion          string `json:"lastDataVersion"`
	LastLoginDate            string `json:"lastLoginDate"`
	LastPlayDate             string `json:"lastPlayDate"`
	LastPlayCredit           int    `json:"lastPlayCredit"`
	LastPlayMode             int    `json:"lastPlayMode"`
	LastPlaceID              int    `json:"lastPlaceId"`
	LastPlaceName            string `json:"lastPlaceName"`
	LastAllNetID             int    `json:"lastAllNetId"`
	LastRegionID             int    `json:"lastRegionId"`
	LastRegionName           string `json:"lastRegionName"`
	LastClientID             string `json:"lastClientId"`
	LastCountryCode          string `json:"lastCountryCode"`
	LastSelectEMoney         int    `json:"lastSelectEMoney"`
	LastSelectTicket         int    `json:"lastSelectTicket"`
	LastSelectCourse         int    `json:"lastSelectCourse"`
	LastCountCourse          int    `json:"lastCountCourse"`
	FirstGameID              string `json:"firstGameId"`
	FirstRomVersion          string `json:"firstRomVersion"`
	FirstDataVersion         string `json:"firstDataVersion"`
	FirstPlayDate            string `json:"firstPlayDate"`
	CompatibleCMVersion      string `json:"compatibleCmVersion"`
	DailyBonusDate           string `json:"dailyBonusDate"`
	DailyCourseBonusDate     string `json:"dailyCourseBonusDate"`
	LastPairLoginDate        string `json:"lastPairLoginDate"`
	LastTrialPlayDate        string `json:"lastTrialPlayDate"`
	PlayVsCount              int    `json:"playVsCount"`
	PlaySyncCount            int    `json:"playSyncCount"`
	WinCount                 int    `json:"winCount"`
	HelpCount                int    `json:"helpCount"`
	ComboCount               int    `json:"comboCount"`
	TotalPoint               int64  `json:"totalPoint"`
	TotalDeluxScore          int64  `json:"totalDeluxscore"`
	TotalBasicDeluxScore     int64  `json:"totalBasicDeluxscore"`
	TotalAdvancedDeluxScore  int64  `json:"totalAdvancedDeluxscore"`
	TotalExpertDeluxScore    int64  `json:"totalExpertDeluxscore"`
	TotalMasterDeluxScore    int64  `json:"totalMasterDeluxscore"`
	TotalReMasterDeluxScore  int64  `json:"totalReMasterDeluxscore"`
	TotalSync                int    `json:"totalSync"`
	TotalBasicSync           int    `json:"totalBasicSync"`
	TotalAdvancedSync        int    `json:"totalAdvancedSync"`
	TotalExpertSync          int    `json:"totalExpertSync"`
	TotalMasterSync          int    `json:"totalMasterSync"`
	TotalReMasterSync        int    `json:"totalReMasterSync"`
	TotalAchievement         int64  `json:"totalAchievement"`
	TotalBasicAchievement    int64  `json:"totalBasicAchievement"`
	TotalAdvancedAchievement int64  `json:"totalAdvancedAchievement"`
	TotalExpertAchievement   int64  `json:"totalExpertAchievement"`
	TotalMasterAchievement   int64  `json:"totalMasterAchievement"`
	TotalReMasterAchievement int64  `json:"totalReMasterAchievement"`
}

type UserOption struct {
	gorm.Model          `json:"-"`
	UserID              int64 `gorm:"uniqueIndex;not null" json:"userId"`
	OptionKind          int   `json:"optionKind"`
	JudgeDisp           int   `json:"judgeDisp"`
	NoteSpeed           int   `json:"noteSpeed"`
	SlideSpeed          int   `json:"slideSpeed"`
	TouchSpeed          int   `json:"touchSpeed"`
	TapDesign           int   `json:"tapDesign"`
	HoldDesign          int   `json:"holdDesign"`
	SlideDesign         int   `json:"slideDesign"`
	StarType            int   `json:"starType"`
	OutlineDesign       int   `json:"outlineDesign"`
	NoteSize            int   `json:"noteSize"`
	SlideSize           int   `json:"slideSize"`
	TouchSize           int   `json:"touchSize"`
	StarRotate          int   `json:"starRotate"`
	DispCenter          int   `json:"dispCenter"`
	DispChain           int   `json:"dispChain"`
	DispRate            int   `json:"dispRate"`
	DispBar             int   `json:"dispBar"`
	TouchEffect         int   `json:"touchEffect"`
	SubmonitorAnimation int   `json:"submonitorAnimation"`
	SubmonitorAchieve   int   `json:"submonitorAchive"`
	SubmonitorAppeal    int   `json:"submonitorAppeal"`
	Matching            int   `json:"matching"`
	TrackSkip           int   `json:"trackSkip"`
	Brightness          int   `json:"brightness"`
	MirrorMode          int   `json:"mirrorMode"`
	DispJudge           int   `json:"dispJudge"`
	DispJudgePos        int   `json:"dispJudgePos"`
	DispJudgeTouchPos   int   `json:"dispJudgeTouchPos"`
	AdjustTiming        int   `json:"adjustTiming"`
	JudgeTiming         int   `json:"judgeTiming"`
	AnsVolume           int   `json:"ansVolume"`
	TapHoldVolume       int   `json:"tapHoldVolume"`
	CriticalSE          int   `json:"criticalSe"`
	TapSE               int   `json:"tapSe"`
	BreakSE             int   `json:"breakSe"`
	BreakVolume         int   `json:"breakVolume"`
	ExSE                int   `json:"exSe"`
	ExVolume            int   `json:"exVolume"`
	SlideSE             int   `json:"slideSe"`
	SlideVolume         int   `json:"slideVolume"`
	TouchHoldVolume     int   `json:"touchHoldVolume"`
	DamageSEVolume      int   `json:"damageSeVolume"`
	HeadPhoneVolume     int   `json:"headPhoneVolume"`
	SortTab             int   `json:"sortTab"`
	SortMusic           int   `json:"sortMusic"`
	OutFrameType        int   `json:"outFrameType"`
	BreakSlideVolume    int   `json:"breakSlideVolume"`
	TouchVolume         int   `json:"touchVolume"`
}

type UserExtend struct {
	gorm.Model                `json:"-"`
	UserID                    int64  `gorm:"uniqueIndex;not null" json:"userId"`
	SelectMusicID             int    `json:"selectMusicId"`
	SelectDifficultyID        int    `json:"selectDifficultyId"`
	CategoryIndex             int    `json:"categoryIndex"`
	MusicIndex                int    `json:"musicIndex"`
	ExtraFlag                 int    `json:"extraFlag"`
	SelectScoreType           int    `json:"selectScoreType"`
	ExtendContentBit          int64  `json:"extendContentBit"`
	IsPhotoAgree              bool   `json:"isPhotoAgree"`
	IsGotoCodeRead            bool   `json:"isGotoCodeRead"`
	SelectResultDetails       bool   `json:"selectResultDetails"`
	SortCategorySetting       int    `json:"sortCategorySetting"`
	SortMusicSetting          int    `json:"sortMusicSetting"`
	PlayStatusSetting         int    `json:"playStatusSetting"`
	SelectResultScoreViewType int    `json:"selectResultScoreViewType"`
	TotalSelect               int    `json:"totalSelect"`
	TotalPlayCount            int    `json:"totalPlayCount"`
	CardID                    int64  `json:"cardId"`
	EventWatchedDate          string `json:"eventWatchedDate"`
}

// UserPlaylog contains the complete persisted fields needed by score history, result views and rating trends.
type UserPlaylog struct {
	gorm.Model            `json:"-"`
	UserID                int64  `gorm:"index:idx_playlog_user_date" json:"userId"`
	OrderID               int64  `json:"orderId"`
	PlaylogID             int64  `json:"playlogId"`
	Version               int    `json:"version"`
	PlaceID               int    `json:"placeId"`
	PlaceName             string `json:"placeName"`
	LoginDate             int64  `json:"loginDate"`
	PlayDate              string `json:"playDate"`
	UserPlayDate          string `gorm:"index:idx_playlog_user_date" json:"userPlayDate"`
	Type                  int    `json:"type"`
	MusicID               int    `json:"musicId"`
	Level                 int    `json:"level"`
	TrackNo               int    `json:"trackNo"`
	Achievement           int    `json:"achievement"`
	DeluxScore            int    `json:"deluxscore"`
	Score                 int    `json:"score"`
	ScoreRank             int    `json:"scoreRank"`
	MaxCombo              int    `json:"maxCombo"`
	TotalCombo            int    `json:"totalCombo"`
	MaxSync               int    `json:"maxSync"`
	TotalSync             int    `json:"totalSync"`
	TapCriticalPerfect    int    `json:"tapCriticalPerfect"`
	TapPerfect            int    `json:"tapPerfect"`
	TapGreat              int    `json:"tapGreat"`
	TapGood               int    `json:"tapGood"`
	TapMiss               int    `json:"tapMiss"`
	HoldCriticalPerfect   int    `json:"holdCriticalPerfect"`
	HoldPerfect           int    `json:"holdPerfect"`
	HoldGreat             int    `json:"holdGreat"`
	HoldGood              int    `json:"holdGood"`
	HoldMiss              int    `json:"holdMiss"`
	SlideCriticalPerfect  int    `json:"slideCriticalPerfect"`
	SlidePerfect          int    `json:"slidePerfect"`
	SlideGreat            int    `json:"slideGreat"`
	SlideGood             int    `json:"slideGood"`
	SlideMiss             int    `json:"slideMiss"`
	TouchCriticalPerfect  int    `json:"touchCriticalPerfect"`
	TouchPerfect          int    `json:"touchPerfect"`
	TouchGreat            int    `json:"touchGreat"`
	TouchGood             int    `json:"touchGood"`
	TouchMiss             int    `json:"touchMiss"`
	BreakCriticalPerfect  int    `json:"breakCriticalPerfect"`
	BreakPerfect          int    `json:"breakPerfect"`
	BreakGreat            int    `json:"breakGreat"`
	BreakGood             int    `json:"breakGood"`
	BreakMiss             int    `json:"breakMiss"`
	IsTap                 bool   `json:"isTap"`
	IsHold                bool   `json:"isHold"`
	IsSlide               bool   `json:"isSlide"`
	IsTouch               bool   `json:"isTouch"`
	IsBreak               bool   `json:"isBreak"`
	IsCriticalDisp        bool   `json:"isCriticalDisp"`
	IsFastLateDisp        bool   `json:"isFastLateDisp"`
	FastCount             int    `json:"fastCount"`
	LateCount             int    `json:"lateCount"`
	IsAchieveNewRecord    bool   `json:"isAchieveNewRecord"`
	IsDeluxScoreNewRecord bool   `json:"isDeluxscoreNewRecord"`
	ComboStatus           int    `json:"comboStatus"`
	SyncStatus            int    `json:"syncStatus"`
	IsClear               bool   `json:"isClear"`
	BeforeRating          int    `json:"beforeRating"`
	AfterRating           int    `json:"afterRating"`
	BeforeGrade           int    `json:"beforeGrade"`
	AfterGrade            int    `json:"afterGrade"`
	AfterGradeRank        int    `json:"afterGradeRank"`
	BeforeDeluxRating     int    `json:"beforeDeluxRating"`
	AfterDeluxRating      int    `json:"afterDeluxRating"`
	IsPlayTutorial        bool   `json:"isPlayTutorial"`
	IsEventMode           bool   `json:"isEventMode"`
	IsFreedomMode         bool   `json:"isFreedomMode"`
	PlayMode              int    `json:"playMode"`
	IsNewFree             bool   `json:"isNewFree"`
	TrialPlayAchievement  int    `json:"trialPlayAchievement"`
	ExtNum1               int    `json:"extNum1"`
	ExtNum2               int    `json:"extNum2"`
	ExtNum4               int    `json:"extNum4"`
	ExtBool1              bool   `json:"extBool1"`
	ExtBool2              bool   `json:"extBool2"`
	ExtBool3              bool   `json:"extBool3"`
	CreateDate            string `json:"createDate"`
}

type UserCharacter struct {
	gorm.Model  `json:"-"`
	UserID      int64 `gorm:"uniqueIndex:idx_user_character;not null" json:"userId"`
	CharacterID int   `gorm:"uniqueIndex:idx_user_character;not null" json:"characterId"`
	Level       int   `json:"level"`
	Count       int   `json:"count"`
	Awakening   int   `json:"awakening"`
	UseCount    int   `json:"useCount"`
}

type UserItem struct {
	gorm.Model `json:"-"`
	UserID     int64 `gorm:"uniqueIndex:idx_user_item;not null" json:"userId"`
	ItemKind   int   `gorm:"uniqueIndex:idx_user_item;not null" json:"itemKind"`
	ItemID     int   `gorm:"uniqueIndex:idx_user_item;not null" json:"itemId"`
	Stock      int   `json:"stock"`
	IsValid    bool  `json:"isValid"`
}

type UserMap struct {
	gorm.Model `json:"-"`
	UserID     int64 `gorm:"uniqueIndex:idx_user_map;not null" json:"userId"`
	MapID      int   `gorm:"uniqueIndex:idx_user_map;not null" json:"mapId"`
	Distance   int   `json:"distance"`
	IsLock     bool  `json:"isLock"`
	IsClear    bool  `json:"isClear"`
	IsComplete bool  `json:"isComplete"`
}

type UserFavorite struct {
	gorm.Model `json:"-"`
	UserID     int64 `gorm:"uniqueIndex:idx_user_favorite;not null" json:"userId"`
	FavUserID  int64 `json:"favUserId"`
	ItemKind   int   `gorm:"uniqueIndex:idx_user_favorite;not null" json:"itemKind"`
	ItemID     int   `json:"itemId"`
	// The database keeps the JSON representation in a text column, while
	// maimai's protocol (and AquaDX) uses an integer array on the wire.
	ItemIDList string `gorm:"type:text" json:"-"`
}

// UnmarshalJSON accepts maimai's itemIdList array and stores its canonical
// JSON representation in the existing text column.
func (u *UserFavorite) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	itemIDList := fields["itemIdList"]
	delete(fields, "itemIdList")
	withoutItemList, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	type favoriteAlias UserFavorite
	if err := json.Unmarshal(withoutItemList, (*favoriteAlias)(u)); err != nil {
		return err
	}
	trimmed := bytes.TrimSpace(itemIDList)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		u.ItemIDList = ""
		return nil
	}
	if trimmed[0] == '[' {
		var values []int
		if err := json.Unmarshal(trimmed, &values); err != nil {
			return err
		}
		canonical, err := json.Marshal(values)
		if err != nil {
			return err
		}
		u.ItemIDList = string(canonical)
		return nil
	}
	if trimmed[0] == '"' {
		return json.Unmarshal(trimmed, &u.ItemIDList)
	}
	return fmt.Errorf("itemIdList must be an array or string")
}

// MarshalJSON restores the array-shaped protocol representation. Older rows
// with an empty or invalid legacy value safely appear as an empty list.
func (u UserFavorite) MarshalJSON() ([]byte, error) {
	values := []int{}
	if len(u.ItemIDList) != 0 {
		_ = json.Unmarshal([]byte(u.ItemIDList), &values)
	}
	type favoriteAlias UserFavorite
	return json.Marshal(struct {
		favoriteAlias
		ItemIDList []int `json:"itemIdList"`
	}{
		favoriteAlias: favoriteAlias(u),
		ItemIDList:    values,
	})
}

type UserMusicDetail struct {
	gorm.Model    `json:"-"`
	UserID        int64 `gorm:"uniqueIndex:idx_user_music_detail;not null" json:"userId"`
	MusicID       int   `gorm:"uniqueIndex:idx_user_music_detail;not null" json:"musicId"`
	Level         int   `gorm:"uniqueIndex:idx_user_music_detail;not null" json:"level"`
	PlayCount     int   `json:"playCount"`
	Achievement   int   `json:"achievement"`
	ComboStatus   int   `json:"comboStatus"`
	SyncStatus    int   `json:"syncStatus"`
	DeluxScoreMax int   `json:"deluxscoreMax"`
	ScoreRank     int   `json:"scoreRank"`
	ExtNum1       int   `json:"extNum1"`
}

type UserCharge struct {
	gorm.Model   `json:"-"`
	UserID       int64  `gorm:"uniqueIndex:idx_user_charge;not null" json:"userId"`
	ChargeID     int    `gorm:"uniqueIndex:idx_user_charge;not null" json:"chargeId"`
	Stock        int    `json:"stock"`
	PurchaseDate string `json:"purchaseDate"`
	ValidDate    string `json:"validDate"`
}

// UserFriendSeasonRanking persists a player's seasonal friend ranking state.
type UserFriendSeasonRanking struct {
	gorm.Model `json:"-"`
	UserID     int64  `gorm:"uniqueIndex:idx_user_friend_season;not null" json:"userId"`
	SeasonID   int    `gorm:"uniqueIndex:idx_user_friend_season;not null" json:"seasonId"`
	Point      int    `json:"point"`
	Rank       int    `json:"rank"`
	RewardGet  bool   `json:"rewardGet"`
	UserName   string `json:"userName"`
	RecordDate string `json:"recordDate"`
}

type UserCourse struct {
	gorm.Model          `json:"-"`
	UserID              int64  `gorm:"uniqueIndex:idx_user_course;not null" json:"userId"`
	CourseID            int    `gorm:"uniqueIndex:idx_user_course;not null" json:"courseId"`
	IsLastClear         bool   `json:"isLastClear"`
	TotalRestLife       int    `json:"totalRestlife"`
	TotalAchievement    int    `json:"totalAchievement"`
	TotalDeluxScore     int    `json:"totalDeluxscore"`
	PlayCount           int    `json:"playCount"`
	ClearDate           string `json:"clearDate"`
	LastPlayDate        string `json:"lastPlayDate"`
	BestAchievement     int    `json:"bestAchievement"`
	BestAchievementDate string `json:"bestAchievementDate"`
	BestDeluxScore      int    `json:"bestDeluxscore"`
	BestDeluxScoreDate  string `json:"bestDeluxscoreDate"`
}

type UserLoginBonus struct {
	gorm.Model `json:"-"`
	UserID     int64 `gorm:"uniqueIndex:idx_user_login_bonus;not null" json:"userId"`
	BonusID    int   `gorm:"uniqueIndex:idx_user_login_bonus;not null" json:"bonusId"`
	Point      int   `json:"point"`
	IsCurrent  bool  `json:"isCurrent"`
	IsComplete bool  `json:"isComplete"`
}

type UserGeneralData struct {
	gorm.Model    `json:"-"`
	UserID        int64  `gorm:"uniqueIndex:idx_user_general_data;not null" json:"userId"`
	PropertyKey   string `gorm:"uniqueIndex:idx_user_general_data;not null" json:"propertyKey"`
	PropertyValue string `gorm:"type:text" json:"propertyValue"`
}

type UserRate struct {
	MusicID     int `json:"musicId"`
	Level       int `json:"level"`
	RomVersion  int `json:"romVersion"`
	Achievement int `json:"achievement"`
}

type UserUdemae struct {
	gorm.Model      `json:"-"`
	UserID          int64 `gorm:"uniqueIndex;not null" json:"userId"`
	Rate            int   `json:"rate"`
	MaxRate         int   `json:"maxRate"`
	ClassValue      int   `json:"classValue"`
	MaxClassValue   int   `json:"maxClassValue"`
	TotalWinNum     int   `json:"totalWinNum"`
	TotalLoseNum    int   `json:"totalLoseNum"`
	MaxWinNum       int   `json:"maxWinNum"`
	MaxLoseNum      int   `json:"maxLoseNum"`
	WinNum          int   `json:"winNum"`
	LoseNum         int   `json:"loseNum"`
	NPCTotalWinNum  int   `json:"npcTotalWinNum"`
	NPCTotalLoseNum int   `json:"npcTotalLoseNum"`
	NPCMaxWinNum    int   `json:"npcMaxWinNum"`
	NPCMaxLoseNum   int   `json:"npcMaxLoseNum"`
	NPCWinNum       int   `json:"npcWinNum"`
	NPCLoseNum      int   `json:"npcLoseNum"`
}

type UserRatingPayload struct {
	Udemae            UserUdemae `json:"udemae"`
	RatingList        []UserRate `json:"ratingList"`
	NewRatingList     []UserRate `json:"newRatingList"`
	NextRatingList    []UserRate `json:"nextRatingList"`
	NextNewRatingList []UserRate `json:"nextNewRatingList"`
}

type UserKaleidx struct {
	gorm.Model          `json:"-"`
	UserID              int64  `gorm:"uniqueIndex:idx_user_kaleidx;not null" json:"userId"`
	GateID              int    `gorm:"uniqueIndex:idx_user_kaleidx;not null" json:"gateId"`
	IsGateFound         bool   `json:"isGateFound"`
	IsKeyFound          bool   `json:"isKeyFound"`
	IsClear             bool   `json:"isClear"`
	TotalRestLife       int    `json:"totalRestLife"`
	TotalAchievement    int    `json:"totalAchievement"`
	TotalDeluxScore     int    `json:"totalDeluxscore"`
	BestAchievement     int    `json:"bestAchievement"`
	BestDeluxScore      int    `json:"bestDeluxscore"`
	BestAchievementDate string `json:"bestAchievementDate"`
	BestDeluxScoreDate  string `json:"bestDeluxscoreDate"`
	PlayCount           int    `json:"playCount"`
	ClearDate           string `json:"clearDate"`
	LastPlayDate        string `json:"lastPlayDate"`
	IsInfoWatched       bool   `json:"isInfoWatched"`
}

type UserIntimate struct {
	gorm.Model            `json:"-"`
	UserID                int64 `gorm:"uniqueIndex:idx_user_intimate;not null" json:"userId"`
	PartnerID             int   `gorm:"uniqueIndex:idx_user_intimate;not null" json:"partnerId"`
	IntimateLevel         int   `json:"intimateLevel"`
	IntimateCountRewarded int   `json:"intimateCountRewarded"`
}

type UserGameCard struct {
	gorm.Model `json:"-"`
	UserID     int64  `gorm:"uniqueIndex:idx_user_game_card;not null" json:"userId"`
	CardID     int    `gorm:"uniqueIndex:idx_user_game_card;not null" json:"cardId"`
	CardTypeID int    `json:"cardTypeId"`
	CharaID    int    `json:"charaId"`
	MapID      int    `json:"mapId"`
	StartDate  string `json:"startDate"`
	EndDate    string `json:"endDate"`
}

type UserPrintDetail struct {
	gorm.Model      `json:"-"`
	UserID          int64  `gorm:"index;not null" json:"userId"`
	UserGameCardID  uint   `json:"-"`
	OrderID         int64  `json:"orderId"`
	PrintNumber     int    `json:"printNumber"`
	PrintDate       string `json:"printDate"`
	SerialID        string `gorm:"uniqueIndex" json:"serialId"`
	PlaceID         int    `json:"placeId"`
	ClientID        string `json:"clientId"`
	PrinterSerialID string `json:"printerSerialId"`
	CardRomVersion  int    `json:"cardRomVersion"`
	IsHolograph     bool   `json:"isHolograph"`
	PrintOption1    bool   `json:"printOption1"`
	PrintOption2    bool   `json:"printOption2"`
	PrintOption3    bool   `json:"printOption3"`
	PrintOption4    bool   `json:"printOption4"`
	PrintOption5    bool   `json:"printOption5"`
	PrintOption6    bool   `json:"printOption6"`
	PrintOption7    bool   `json:"printOption7"`
	PrintOption8    bool   `json:"printOption8"`
	PrintOption9    bool   `json:"printOption9"`
	PrintOption10   bool   `json:"printOption10"`
	Created         string `json:"created"`
}

// UserRegion tracks a player's play count in a game region, matching AquaDX UserRegions.
type UserRegion struct {
	gorm.Model `json:"-"`
	UserID     int64 `gorm:"uniqueIndex:idx_user_region;not null" json:"userId"`
	RegionID   int   `gorm:"uniqueIndex:idx_user_region;not null" json:"regionId"`
	PlayCount  int   `json:"playCount"`
}

type UserActivity struct {
	gorm.Model `json:"-"`
	UserID     int64 `gorm:"index;not null" json:"userId"`
	Kind       int   `json:"kind"`
	ActivityID int   `json:"id"`
	SortNumber int64 `json:"sortNumber"`
	Param1     int   `json:"param1"`
	Param2     int   `json:"param2"`
	Param3     int   `json:"param3"`
	Param4     int   `json:"param4"`
}

type UserActivityPayload struct {
	PlayList  []UserActivity `json:"playList"`
	MusicList []UserActivity `json:"musicList"`
}

type UserFavoriteMusic struct {
	OrderID int `json:"orderId"`
	ID      int `json:"id"`
}

// UpsertUserAllRequest matches the collection-based payload accepted by UpsertUserAllApi.
type UpsertUserAllRequest struct {
	UserID        int64 `json:"userId"`
	UpsertUserAll struct {
		UserData                    []UserDetail              `json:"userData"`
		UserOption                  []UserOption              `json:"userOption"`
		UserExtend                  []UserExtend              `json:"userExtend"`
		UserCharacterList           []UserCharacter           `json:"userCharacterList"`
		UserMapList                 []UserMap                 `json:"userMapList"`
		UserLoginBonusList          []UserLoginBonus          `json:"userLoginBonusList"`
		UserRatingList              []UserRatingPayload       `json:"userRatingList"`
		UserItemList                []UserItem                `json:"userItemList"`
		UserMusicDetailList         []UserMusicDetail         `json:"userMusicDetailList"`
		UserCourseList              []UserCourse              `json:"userCourseList"`
		UserChargeList              []UserCharge              `json:"userChargeList"`
		UserFriendSeasonRankingList []UserFriendSeasonRanking `json:"userFriendSeasonRankingList"`

		UserFavoriteList       []UserFavorite        `json:"userFavoriteList"`
		UserActivityList       []UserActivityPayload `json:"userActivityList"`
		UserFavoriteMusicList  []UserFavoriteMusic   `json:"userFavoritemusicList"`
		UserKaleidxScopeList   []UserKaleidx         `json:"userKaleidxScopeList"`
		UserIntimateList       []UserIntimate        `json:"userIntimateList"`
		UserPlaylogList        []UserPlaylog         `json:"userPlaylogList"`
		IsNewFavoriteMusicList string                `json:"isNewFavoritemusicList"`
	} `json:"upsertUserAll"`
	UserPlaylogList []UserPlaylog `json:"userPlaylogList"`
}
