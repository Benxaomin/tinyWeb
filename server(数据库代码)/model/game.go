// Package model 游戏相关数据模型
// =============================================
// 作用：
//   定义游戏排行榜的数据结构
//
// FlappyBirdScore - Flappy Bird游戏得分记录
// =============================================
package model

import (
	"time"

	"gorm.io/gorm"
)

// FlappyBirdScore Flappy Bird游戏得分记录
// 对应数据库 flappy_bird_scores 表
//
// 设计思路：
//   - 只记录登录用户的得分（匿名用户不记录）
//   - 记录得分、游戏时间、创建时间
//   - 通过 user_id 关联用户表
//
// 数据库表结构：
//
//	| 列名       | 类型          | 说明                        |
//	|------------|---------------|----------------------------|
//	| id         | bigint unsigned| 自增主键                    |
//	| created_at | datetime(3)   | 创建时间                    |
//	| updated_at | datetime(3)   | 更新时间                    |
//	| deleted_at | datetime(3)   | 软删除时间                  |
//	| user_id    | bigint unsigned| 关联用户ID（索引）           |
//	| score      | int           | 游戏得分                    |
//	| game_time  | int           | 游戏时长（秒）              |
//	| played_at  | datetime(3)   | 游戏进行时间                |
type FlappyBirdScore struct {
	gorm.Model
	UserID   uint      `gorm:"index;not null" json:"user_id"`   // 关联用户ID
	Score    int       `gorm:"not null" json:"score"`           // 游戏得分
	GameTime int       `gorm:"default:0" json:"game_time"`      // 游戏时长（秒）
	PlayedAt time.Time `gorm:"not null" json:"played_at"`       // 游戏进行时间
}

// TableName 指定 FlappyBirdScore 对应的数据库表名
func (FlappyBirdScore) TableName() string {
	return "flappy_bird_scores"
}

// ---- API 请求/响应结构体 ----

// SaveScoreRequest 保存游戏得分的请求体
// 前端 POST /api/game/flappy-bird/score 时提交的 JSON 数据
type SaveScoreRequest struct {
	Score    int `json:"score" binding:"required,min=0"`     // 游戏得分（必填，≥0）
	GameTime int `json:"game_time"`                          // 游戏时长（秒，可选）
}

// ScoreItem 排行榜单项
type ScoreItem struct {
	Username  string    `json:"username"`   // 用户名
	Score     int       `json:"score"`      // 得分
	GameTime  int       `json:"game_time"`  // 游戏时长（秒）
	PlayedAt  time.Time `json:"played_at"`  // 游戏时间
	Rank      int       `json:"rank"`       // 排名
}

// LeaderboardResponse 排行榜响应
// GET /api/game/flappy-bird/leaderboard 返回的数据
type LeaderboardResponse struct {
	GlobalTop10 []ScoreItem `json:"global_top10"` // 全服Top10
	MyTop5      []ScoreItem `json:"my_top5"`      // 个人Top5
	MyBestScore int         `json:"my_best_score"` // 个人最高分
	TotalPlays  int64       `json:"total_plays"`   // 总游戏次数
	IsLoggedIn  bool        `json:"is_logged_in"`  // 是否已登录（新增字段）
}

// PersonalScoreResponse 个人得分记录响应
// GET /api/game/flappy-bird/my-scores 返回的数据
type PersonalScoreResponse struct {
	Scores     []ScoreItem `json:"scores"`      // 个人得分列表
	BestScore  int         `json:"best_score"`  // 最高分
	TotalGames int64       `json:"total_games"` // 总游戏次数
	TotalTime  int         `json:"total_time"`  // 总游戏时长（秒）
}
