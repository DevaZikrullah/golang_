package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"test/logging"
	"test/middleware"
	"test/models"
	"test/utils"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
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
		logging.Error("Failed to read request body", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		logging.Error("Invalid request body", zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	emailRegex := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	if !emailRegex.MatchString(input.Email) {
		logging.Warn("Email must be in valid format")
		utils.RespondWithError(w, http.StatusBadRequest, "Email must be in valid format")
		return
	}

	existingUser := models.Users{}
	err = models.DB.Where("email = ?", input.Email).First(&existingUser).Error
	if err == nil {
		logging.Warn("Email already exists")
		utils.RespondWithError(w, http.StatusConflict, "Email already exists")
		return
	}

	err = models.DB.Where("username = ?", input.Username).First(&existingUser).Error
	if err == nil {
		logging.Warn("Username already exists")
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
		logging.Error("Failed to create user", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	logging.Info("User created", zap.String("username", newUser.Username), zap.String("email", newUser.Email))
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
		logging.Error("Failed to read request body", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		logging.Error("Invalid request body", zap.Error(err))
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
		logging.Warn("User not found")
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if existingUser.Password != input.Password {
		logging.Warn("Invalid Credentials")
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := generateJWTToken(existingUser.ID)
	if err != nil {
		logging.Warn("Failed Generate token")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	existingUser.Token = token

	models.DB.Save(&existingUser)

	logging.Info("User Login", zap.String("username", existingUser.Username), zap.String("token", existingUser.Token))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func GetInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIdFromToken(r)
	if err != nil {
		logging.Warn("Unauthorized")
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.Users
	if err := models.DB.Preload("Quests").Preload("CompletedQuests").Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logging.Warn("User not found")
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

	logging.Info("User Generate Token", zap.String("userID", strconv.Itoa(int(userID))))

	secretKey := []byte("your-secret-key")
	return token.SignedString(secretKey)
}
