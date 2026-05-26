package models

import (
	"time"
)

type User struct {
	ID           		uint      `gorm:"primaryKey"`
	Username     		string    `gorm:"type:varchar(20);not null;unique"`
	Email        		string    `gorm:"type:varchar(255);not null;unique"`
	Password     		string    `gorm:"type:varchar(60);not null"`
	RefreshToken 		string    `gorm:"type:text"`
	IsAdmin      		bool      `gorm:"default:false"`
	IsOnline     		bool      `gorm:"default:false"`
	CreatedAt    		time.Time `gorm:"autoCreateTime"`
	TwoFactorCode       string    `gorm:"size:6"`
	TwoFactorExpiresAt  time.Time `gorm:"default:null"`
	IsActive       		bool      `gorm:"default:false"`
}

type Post struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"type:varchar(255);not null"`
	Content   string    `gorm:"column:content_post;type:text;not null"`
	ImageUrl  string    `gorm:"type:varchar(255)"`
	UserID    uint      `gorm:"not null"`
	Status    string    `gorm:"type:varchar(20);default:'pending'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Tags      []Tag     `gorm:"many2many:post_tags;"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(27);not null;unique"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey"`
	Content   string    `gorm:"type:text;not null"`
	UserID    uint      `gorm:"not null"`
	PostID    uint      `gorm:"not null"`
	Status    string    `gorm:"type:varchar(20);default:'pending'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type Like struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"not null;uniqueIndex:idx_user_post"`
	PostID uint `gorm:"not null;uniqueIndex:idx_user_post"`
	IsLike bool `gorm:"not null"`
}

type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}

type UserToken struct {
	ID		uint    	`gorm:"primaryKey"`
	UserID	uint    	`gorm:"not null"`
	Token	string  	`gorm:"type:text;not null"`
	Type	string    	`gorm:"type:varchar(20);not null"` // "access" ou "refresh"
	ExpiresAt time.Time `gorm:"not null"`
}