package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"test/logging"
	"test/middleware"
	"test/models"
	"test/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// var validate *validator.Validate

func GetAllUom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var uom []models.Uom
	models.DB.Find(&uom)

	json.NewEncoder(w).Encode(uom)
}

func GetUom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	var uom models.Uom

	if err := models.DB.Where("id = ?", id).First(&uom).Error; err != nil {
		logging.Warn("Uom not found")
		utils.RespondWithError(w, http.StatusNotFound, "Uom not found")
		return
	}

	json.NewEncoder(w).Encode(uom)
}

type UomInput struct {
	Name string `json:"name" validate:"required"`
}

func CreateUom(w http.ResponseWriter, r *http.Request) {
	var input UomInput

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

	uom := &models.Uom{
		Name:   input.Name,
		UserID: userID,
	}

	err = models.DB.Create(uom).Error
	if err != nil {
		logging.Error(err.Error(), zap.Error(err))
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create uom")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uom)
}

func UpdateUom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	var uom models.Uom

	if err := models.DB.Where("id = ?", id).First(&uom).Error; err != nil {
		logging.Warn("uom not found")
		utils.RespondWithError(w, http.StatusNotFound, "uom not found")
		return
	}

	var input UomInput

	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &input)

	validate = validator.New()
	err := validate.Struct(input)

	if err != nil {
		logging.Error(err.Error(), zap.Error(err))
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	uom.Name = input.Name

	models.DB.Save(&uom)

	json.NewEncoder(w).Encode(uom)
}

func DeleteUom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := mux.Vars(r)["id"]
	var uom models.Uom

	if err := models.DB.Where("id = ?", id).First(&uom).Error; err != nil {
		logging.Warn("Uom not fouund")
		utils.RespondWithError(w, http.StatusNotFound, "Uom not found")
		return
	}

	models.DB.Delete(&uom)

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(uom)
}
