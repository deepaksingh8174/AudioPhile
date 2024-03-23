package utils

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
)

type Status struct {
	Message string `json:"Message"`
}

func RespondJSON (w http.ResponseWriter,statusCode int, body interface{}) {
	w.Header().Set("content-type","application/json; charset = utf-8")
	w.WriteHeader(statusCode)
	encodeErr := json.NewEncoder(w).Encode(body)
	if encodeErr != nil {
		logrus.Errorf("Failed to encode the message")
	}
}


func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateSession() *session.Session {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String(os.Getenv("region")),
			Credentials: credentials.NewStaticCredentials(
				os.Getenv("accessKey"),
				os.Getenv("secret_key"),
				"",
			),
		},
	))
	return sess
}


func CreateS3Session(sess *session.Session) *s3.S3 {
	s3Session := s3.New(sess)
	return s3Session
}