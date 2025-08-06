package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
var mySigningKey = []byte(JWT_SECRET_KEY)

type MyCustomClaims struct {
	ID     uuid.UUID   `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	RoleID uuid.UUID   `json:"role_id"`
	jwt.RegisteredClaims
}

func CreateToken(user *models.User) (string, error) {
	var roleID uuid.UUID
	if user.RoleID != nil {
		roleID = *user.RoleID
	} else {
		return "", fmt.Errorf("RoleID is nil")
	}

	claims := MyCustomClaims{
		ID:     user.ID,
		Name:   user.Name,
		Email:  user.Email,
		RoleID: roleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", err	
	}

	return ss, nil
}

func ValidateToken(tokenString string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (any, error) {
		return mySigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*MyCustomClaims)
	
	if !ok || !token.Valid {
		return nil, err
	}
	
	return claims, nil
}
