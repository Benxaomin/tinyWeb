// Package handler 游戏相关API处理器
// =============================================
// 作用：
//
//	处理游戏排行榜相关的HTTP请求
//
// 路由：
//
//	POST   /api/game/flappy-bird/score      - 保存游戏得分
//	GET    /api/game/flappy-bird/leaderboard - 获取排行榜（全服Top10 + 个人Top5）
//	GET    /api/game/flappy-bird/my-scores   - 获取个人所有得分记录
//
// =============================================
package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"tinyweb1/middleware"
	"tinyweb1/model"

	"gorm.io/gorm"
)

// GameHandler 游戏处理器结构体
type GameHandler struct {
	DB *gorm.DB
}

// NewGameHandler 创建游戏处理器实例
func NewGameHandler(db *gorm.DB) *GameHandler {
	return &GameHandler{DB: db}
}

// ============================================
// HTTP适配器函数：将handler函数转换为标准库格式
// ============================================

// SaveScoreHandler 保存得分HTTP处理器
func SaveScoreHandler(h *GameHandler, w http.ResponseWriter, r *http.Request) {
	// 从JWT上下文中获取用户ID（由中间件设置）
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		sendJSON(w, http.StatusUnauthorized, model.ErrorResponse(401, "请先登录"))
		return
	}

	var req model.SaveScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, model.ErrorResponse(400, "参数错误："+err.Error()))
		return
	}

	// 创建得分记录
	score := model.FlappyBirdScore{
		UserID:   userID.(uint),
		Score:    req.Score,
		GameTime: req.GameTime,
		PlayedAt: time.Now(),
	}

	if err := h.DB.Create(&score).Error; err != nil {
		sendJSON(w, http.StatusInternalServerError, model.ErrorResponse(500, "保存得分失败"))
		return
	}

	sendJSON(w, http.StatusOK, model.SuccessResponse(map[string]interface{}{
		"message": "得分已保存",
		"score":   req.Score,
	}))
}

// GetLeaderboardHandler 获取排行榜HTTP处理器（公开接口，无需登录）
func GetLeaderboardHandler(h *GameHandler, w http.ResponseWriter, r *http.Request) {
	// 从JWT上下文中获取用户ID（可能未登录）
	userID := r.Context().Value(middleware.UserIDKey)
	uid := uint(0)
	isLoggedIn := false
	if userID != nil {
		uid = userID.(uint)
		isLoggedIn = true
	}

	// 1. 获取全服Top10 - 每个用户只显示最好成绩（取最早达成的最高分记录）
	var globalTop10 []model.ScoreItem
	err := h.DB.Raw(`
		SELECT 
			u.username,
			fbs.score,
			fbs.game_time,
			fbs.played_at,
			0 as ` + "`rank`" + `
		FROM (
			SELECT user_id, MAX(score) as max_score
			FROM flappy_bird_scores
			GROUP BY user_id
		) best
		JOIN flappy_bird_scores fbs 
			ON best.user_id = fbs.user_id 
			AND best.max_score = fbs.score
		JOIN users u ON fbs.user_id = u.id
		WHERE fbs.played_at = (
			SELECT MIN(played_at)
			FROM flappy_bird_scores
			WHERE user_id = fbs.user_id AND score = fbs.score
		)
		ORDER BY fbs.score DESC, fbs.played_at ASC
		LIMIT 10
	`).Scan(&globalTop10).Error

	if err != nil {
		globalTop10 = []model.ScoreItem{}
	}

	// 手动设置排名
	for i := range globalTop10 {
		globalTop10[i].Rank = i + 1
	}

	// 2. 获取个人Top5（仅登录用户）
	var myTop5 []model.ScoreItem
	if isLoggedIn {
		// 先获取用户所有记录，按分数降序
		err = h.DB.Raw(`
			SELECT 
				u.username,
				fbs.score,
				fbs.game_time,
				fbs.played_at,
				0 as `+"`rank`"+`
			FROM flappy_bird_scores fbs
			JOIN users u ON fbs.user_id = u.id
			WHERE fbs.user_id = ?
			ORDER BY fbs.score DESC, fbs.played_at DESC
			LIMIT 5
		`, uid).Scan(&myTop5).Error

		if err != nil {
			myTop5 = []model.ScoreItem{}
		}

		// 手动设置排名
		for i := range myTop5 {
			myTop5[i].Rank = i + 1
		}
	}

	// 3. 获取个人最高分（仅登录用户）
	var myBestScore int
	if isLoggedIn {
		h.DB.Model(&model.FlappyBirdScore{}).
			Where("user_id = ?", uid).
			Select("COALESCE(MAX(score), 0)").
			Scan(&myBestScore)
	}

	// 4. 获取总游戏次数
	var totalPlays int64
	h.DB.Model(&model.FlappyBirdScore{}).Count(&totalPlays)

	sendJSON(w, http.StatusOK, model.SuccessResponse(model.LeaderboardResponse{
		GlobalTop10: globalTop10,
		MyTop5:      myTop5,
		MyBestScore: myBestScore,
		TotalPlays:  totalPlays,
		IsLoggedIn:  isLoggedIn,
	}))
}

// GetMyScoresHandler 获取个人所有得分HTTP处理器
func GetMyScoresHandler(h *GameHandler, w http.ResponseWriter, r *http.Request) {
	// 从JWT上下文中获取用户ID
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		sendJSON(w, http.StatusUnauthorized, model.ErrorResponse(401, "请先登录"))
		return
	}
	uid := userID.(uint)

	// 获取个人所有得分（按得分降序，最多50条）
	var scores []model.ScoreItem
	err := h.DB.Raw(`
		SELECT 
			u.username,
			fbs.score,
			fbs.game_time,
			fbs.played_at,
			0 as `+"`rank`"+`
		FROM flappy_bird_scores fbs
		JOIN users u ON fbs.user_id = u.id
		WHERE fbs.user_id = ?
		ORDER BY fbs.score DESC, fbs.played_at ASC
		LIMIT 50
	`, uid).Scan(&scores).Error

	if err != nil {
		scores = []model.ScoreItem{}
	}

	// 手动设置排名
	for i := range scores {
		scores[i].Rank = i + 1
	}

	// 获取统计数据
	var bestScore int
	var totalGames int64
	var totalTime int

	h.DB.Model(&model.FlappyBirdScore{}).
		Where("user_id = ?", uid).
		Select("COALESCE(MAX(score), 0)").
		Scan(&bestScore)

	h.DB.Model(&model.FlappyBirdScore{}).
		Where("user_id = ?", uid).
		Count(&totalGames)

	h.DB.Model(&model.FlappyBirdScore{}).
		Where("user_id = ?", uid).
		Select("COALESCE(SUM(game_time), 0)").
		Scan(&totalTime)

	sendJSON(w, http.StatusOK, model.SuccessResponse(model.PersonalScoreResponse{
		Scores:     scores,
		BestScore:  bestScore,
		TotalGames: totalGames,
		TotalTime:  totalTime,
	}))
}
