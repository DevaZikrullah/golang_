package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"test/middleware"
	"test/models"
	"test/utils"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

type UserInput struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var input UserInput

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	emailRegex := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	if !emailRegex.MatchString(input.Email) {
		utils.RespondWithError(w, http.StatusBadRequest, "Email must be in valid format")
		return
	}

	existingUser := models.Users{}
	err = models.DB.Where("email = ?", input.Email).First(&existingUser).Error
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "Email already exists")
		return
	}

	err = models.DB.Where("username = ?", input.Username).First(&existingUser).Error
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "Username already exists")
		return
	}

	newUser := &models.Users{
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}
	err = models.DB.Create(newUser).Error
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newUser)
}

type LoginInput struct {
	LoginIdentifier string `json:"login_identifier" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	existingUser := models.Users{}

	identifier := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	var userQuery *gorm.DB
	if !identifier.MatchString(input.LoginIdentifier) {
		userQuery = models.DB.Where("username = ?", input.LoginIdentifier)
	} else {
		userQuery = models.DB.Where("email = ?", input.LoginIdentifier)
	}

	if err := userQuery.First(&existingUser).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if existingUser.Password != input.Password {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := generateJWTToken(existingUser.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	existingUser.Token = token

	models.DB.Save(&existingUser)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func GetInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIdFromToken(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.Users
	if err := models.DB.Preload("Quests").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	json.NewEncoder(w).Encode(user)
}

func generateJWTToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(time.Duration(6) * time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    expirationTime.Unix(),
	})

	secretKey := []byte("your-secret-key")
	return token.SignedString(secretKey)
}
