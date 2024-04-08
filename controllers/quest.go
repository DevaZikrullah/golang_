package controllers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"test/logging"
	"test/middleware"
	"test/models"
	"test/utils"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var validate *validator.Validate

func GetAllQuests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var quests []models.Quest
	models.DB.Find(&quests)

	json.NewEncoder(w).Encode(quests)
}

func GetQuest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	var quest models.Quest

	if err := models.DB.Where("id = ?", id).First(&quest).Error; err != nil {
		logging.Warn("Quest not found")
		utils.RespondWithError(w, http.StatusNotFound, "Quest not found")
		return
	}

	json.NewEncoder(w).Encode(quest)
}

type QuestInput struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
	Reward      int    `json:"reward" validate:"required"`
}

func CreateQuest(w http.ResponseWriter, r *http.Request) {
	var input QuestInput

	userID, err := middleware.GetUserIdFromToken(r)
	if err != nil {
		logging.Error("Unauthorized", zap.Error(err))
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logging.Error("Internal Server Error", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		logging.Error("Invalid request body", zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	validate = validator.New()
	err = validate.Struct(input)
	if err != nil {
		logging.Error(err.Error(), zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	quest := &models.Quest{
		Title:       input.Title,
		Description: input.Description,
		Reward:      input.Reward,
		UserID:      userID,
	}

	err = models.DB.Create(quest).Error
	if err != nil {
		logging.Error(err.Error(), zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create quest")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quest)
}

func UpdateQuest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	var quest models.Quest

	if err := models.DB.Where("id = ?", id).First(&quest).Error; err != nil {
		logging.Warn("Quest not found")
		utils.RespondWithError(w, http.StatusNotFound, "Quest not found")
		return
	}

	var input QuestInput

	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &input)

	validate = validator.New()
	err := validate.Struct(input)

	if err != nil {
		logging.Error(err.Error(), zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	quest.Title = input.Title
	quest.Description = input.Description
	quest.Reward = input.Reward

	models.DB.Save(&quest)

	json.NewEncoder(w).Encode(quest)
}

func DeleteQuest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	var quest models.Quest

	if err := models.DB.Where("id = ?", id).First(&quest).Error; err != nil {
		logging.Warn("Quest not fouund")
		utils.RespondWithError(w, http.StatusNotFound, "Quest not found")
		return
	}

	models.DB.Delete(&quest)

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(quest)
}

type InputQuestComplete struct {
	QuestId int `json:"quest_id" validate:"required"`
	UserId  int `json:"user_id" validate:"required"`
}

func QuestComplete(w http.ResponseWriter, r *http.Request) {
	var input InputQuestComplete
	var quest models.Quest
	var user models.Users

	userID, err := middleware.GetUserIdFromToken(r)
	if err != nil {
		logging.Error("Unauthorized", zap.Error(err))
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logging.Error("Internal Server Error", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = json.Unmarshal(body, &input)
	if err != nil {
		logging.Error("Invalid Json", zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	validate = validator.New()
	err = validate.Struct(input)
	if err != nil {
		logging.Error("Validation Erorr", zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	if err := models.DB.Where("id = ?", input.QuestId).First(&quest).Error; err != nil {
		logging.Warn("Quest not found")
		utils.RespondWithError(w, http.StatusNotFound, "Quest not found")
		return
	}

	if err := models.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		logging.Warn("User not found")
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	user.Point += quest.Reward

	completeQuest := &models.CompletedQuest{
		UserID:      userID,
		QuestID:     quest.ID,
		CompletedAt: time.Now(),
	}

	user.AppendCompletedQuest(*completeQuest)
	if err := models.DB.Save(&user).Error; err != nil {
		logging.Error("Failed to update user", zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
