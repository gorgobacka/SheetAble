package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/SheetAble/SheetAble/backend/api/auth"
	"github.com/SheetAble/SheetAble/backend/api/models"
	"github.com/SheetAble/SheetAble/backend/api/responses"
	"github.com/gorilla/mux"
)

func (server *Server) GetSheetsPage(w http.ResponseWriter, r *http.Request) {
	/*
		This endpoint will return all sheets in Page like style.
		Meaning POST request will have 3 attributes:
			- sort_by: (how is it sorted)
			- page: (what page)
			- limit: (limit number)
			- composer: (what composer)

		Return:
			- sheets: [...]
			- page_max: [7] // How many pages there are
			- page_current: [1] // Which page is currently selected
	*/

	sortBy := r.FormValue("sort_by")
	if sortBy == "" {
		sortBy = "updated_at desc"
	}

	limitInt := 0
	limit := r.FormValue("limit")
	if limit == "" {
		limitInt = 10
	} else {
		limitInt, _ = strconv.Atoi(limit)
	}

	pageInt := 0
	page := r.FormValue("page")
	if page == "" {
		pageInt = 1
	} else {
		pageInt, _ = strconv.Atoi(page)
	}

	pagination := models.Pagination{
		Sort:  sortBy,
		Limit: limitInt,
		Page:  pageInt,
	}

	sheet := models.Sheet{}
	pageNew, _ := sheet.List(server.DB, pagination, r.FormValue("composer"))

	responses.JSON(w, http.StatusOK, pageNew)
}

func (server *Server) GetSheet(w http.ResponseWriter, r *http.Request) {
	/*
		Get PDF file and information about an individual sheet.
		Example request: /sheet/Étude N. 1
		Has to be safeName
	*/

	vars := mux.Vars(r)
	sheetName := vars["sheetName"]
	if sheetName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("missing URL parameter 'sheetName'"))
		return
	}

	sheetModel := models.Sheet{}
	fmt.Println(sheetName)

	sheet, _ := sheetModel.FindSheetBySafeName(server.DB, sheetName)

	responses.JSON(w, http.StatusOK, sheet)
}

func (server *Server) GetPDF(w http.ResponseWriter, r *http.Request) {
	/*
		Serve the PDF file
		Example request: /sheet/pdf/Frédéric Chopin/Étude N. 1
		sheetname and composer name have to be the safeName of them
	*/

	name := mux.Vars(r)["sheetName"]
	composer := mux.Vars(r)["composer"]
	http.ServeFile(w, r, os.Getenv("CONFIG_PATH")+"sheets/uploaded-sheets/"+composer+"/"+name+".pdf")
}

func (server *Server) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	/*
		Serve the thumbnail file
		name = safename of sheet
	*/

	name := mux.Vars(r)["name"]
	http.ServeFile(w, r, os.Getenv("CONFIG_PATH")+"sheets/thumbnails/"+name+".png")
}

func (server *Server) DeletSheet(w http.ResponseWriter, r *http.Request) {
	/*
		Has to be safeName of the sheet
	*/

	vars := mux.Vars(r)
	sheetName := vars["sheetName"]

	// Is this user authenticated?
	_, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// Check if the sheet exist
	sheet := models.Sheet{}
	err = server.DB.Model(models.Sheet{}).Where("safe_sheet_name = ?", sheetName).Take(&sheet).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("sheet not found"))
		return
	}

	_, err = sheet.DeleteSheet(server.DB, sheetName)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, "Sheet was successfully deleted")
}
