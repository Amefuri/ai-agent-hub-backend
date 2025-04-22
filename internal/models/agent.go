package models

import (
	"gorm.io/gorm"
)

type Agent struct {
	gorm.Model
	Name          string `json:"name"`
	Description   string `json:"description"`
	Avatar        string `json:"avatar"`
	SystemPrompt  string `json:"system_prompt"`
	InputTemplate string `json:"input_template"`
	Personality   string `json:"personality"`
	UserID        uint   `json:"userId"`
	IsFeatured    bool   `json:"isFeatured"`
	ViewCount     uint   `json:"viewCount"`
}
