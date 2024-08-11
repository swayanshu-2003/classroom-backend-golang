package models

import (
	"time"

	"gorm.io/gorm"
)

// Users represents the user model
type Users struct {
	// ID             int                     `gorm:"primaryKey;autoIncrement" json:"id"`
	Uuid           *string                 `gorm:"primaryKey" json:"uuid"`
	Username       *string                 `json:"username"`
	Name           *string                 `json:"name"`
	Password       *string                 `json:"password"`
	ProfilePicture *string                 `json:"profile_picture"`
	Classroom      []Classroom             `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"classroom"`
	Collaborations []ClassroomCollaborator `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"collaborations"`
	Comments       []Comment               `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE" json:"comments"`
	Assignments    []Assignments           `gorm:"foreignKey:AutherId;constraint:OnDelete:CASCADE" json:"assignments"`
}

// Classroom represents the classroom model
type Classroom struct {
	// ID            int                     `gorm:"primaryKey;autoIncrement" json:"id"`
	ClassId       *string                 `gorm:"primaryKey" json:"class_id"`
	ClassName     *string                 `json:"class_name"`
	Description   *string                 `json:"description"`
	Done          bool                    `gorm:"default:false" json:"done"`
	OwnerID       *string                 `json:"owner_id"`
	IsDeleted     bool                    `gorm:"default:false" json:"is_deleted"`
	Shared        bool                    `gorm:"default:false" json:"shared"`
	Owner         Users                   `gorm:"foreignKey:OwnerID;references:Uuid" json:"owner"`
	Collaborators []ClassroomCollaborator `gorm:"foreignKey:ClassID;constraint:OnDelete:CASCADE" json:"collaborators"`
	Comments      []Comment               `gorm:"foreignKey:ClassID;constraint:OnDelete:CASCADE" json:"comments"`
	CreatedAt     time.Time               `gorm:"default:now()" json:"created_at"`
	Assignments   []Assignments           `gorm:"foreignKey:ClassID;constraint:OnDelete:CASCADE" json:"assignments"`
}

// ClassroomCollaborator represents the classroom collaborator model
type ClassroomCollaborator struct {
	UserID    *string `json:"user_id"`
	ClassID   *string `json:"class_id"`
	Role      string  `json:"role"`
	IsRemoved bool    `gorm:"default:false" json:"is_removed"`
	// User      Users     `gorm:"-" json:"-"`
	// Classroom Classroom `gorm:"-" json:"-"`
	User      Users     `gorm:"foreignKey:UserID;references:Uuid" json:"users"`
	Classroom Classroom `gorm:"foreignKey:ClassID;references:ClassId" json:"classroom"`
}

// Comment represents the comment model
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   *string   `json:"content"`
	AuthorID  *string   `json:"author_id"`
	ClassID   *string   `json:"class_id"`
	CreatedAt time.Time `gorm:"default:now()" json:"created_at"`
	Author    Users     `gorm:"foreignKey:AuthorID;references:Uuid" json:"author"`
	Classroom Classroom `gorm:"foreignKey:ClassID;references:ClassId" json:"classroom"`
}

// assignments
type Assignments struct {
	IsDeleted   bool      `gorm:"default:false" json:"is_deleted"`
	ID          *string   `gorm:"primaryKey" json:"id"`
	Title       *string   `json:"title"`
	Type        *string   `json:"type"`
	Description *string   `json:"description"`
	Link        *string   `json:"link"`
	ClassID     *string   `json:"class_id"`
	AutherId    *string   `json:"auther_id"`
	CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
	Classroom   Classroom `gorm:"foreignKey:ClassID;references:ClassId" json:"classroom"`
	CreatedBy   Users     `gorm:"foreignKey:AutherId;references:Uuid" json:"created_by"`
}

// MigrateUser migrates the user and related models
func MigrateUser(db *gorm.DB) error {
	err := db.AutoMigrate(&Users{}, &Classroom{}, &ClassroomCollaborator{}, &Comment{}, &Assignments{})
	return err
}
