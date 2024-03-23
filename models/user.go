package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

const (
	AdminRole = "admin"
	UserRole = "user"
	HeadPhoneType = "headphone"
	EarphoneType = "earphone"
	SpeakerType = "speaker"
)

type User struct {
	Name string  `json:"name"`
	Email string  `json:"email"`
	Password string `json:"password"`
	Phone string `json:"phone"`

}

type Login struct {
	Email string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type Claims struct {
	UserId uuid.UUID `json:"userid"`
	jwt.StandardClaims

}

type Address struct {
	Id uuid.UUID `db:"id"`
	Name string `json:"name" db:"name"`
	PinCode string `json:"pin_code" db:"pin_code"`
	Latitude float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
}

type Users struct {
	Id  uuid.UUID `db:"id"`
	Name string `json:"name" db:"name"`
	Email string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
	Phone string `json:"phone" db:"phone"`
	CreatedAt time.Time `db:"created_at"`
	//ArchivedAt time.Time `db:"archived_at"`
}

type Product struct {
	Name string `json:"name" db:"name"`
	Cost float64 `json:"cost" db:"cost"`
	Type string `json:"type" db:"type"`
	Quantity int `json:"quantity" db:"quantity"`
}

type Products struct {
	Id uuid.UUID         `db:"id"`
	Name string          `json:"name" db:"name"`
	Cost int             `json:"cost" db:"cost"`
	Type string          `json:"type" db:"item_type"`
	Quantity int         `json:"quantity" db:"quantity"`
	Image  pq.StringArray      `db:"images"`
}

type Upload struct {
	Id uuid.UUID    `db:"id"`
	Path string     `db:"path"`
	FileName string  `db:"name"`
	Url string   `json:"url" db:"url"`
}

type CartItems struct {
	ProductId uuid.UUID  `db:"product_id"`
	Quantity int    `db:"quantity"`

}

type QuantityLeft struct {
	ProductId uuid.UUID
	QuantityLeft int
}

type AWSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	BucketName      string
}
