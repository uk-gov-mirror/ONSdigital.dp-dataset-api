package api

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/go-ns/audit"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

func (api *DatasetAPI) getEditions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	logData := log.Data{"dataset_id": id}
	auditParams := common.Params{"dataset_id": id}

	if auditErr := api.auditor.Record(r.Context(), getEditionsAction, audit.Attempted, auditParams); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	authorised, logData := api.authenticate(r, logData)

	var state string
	if !authorised {
		state = models.PublishedState
	}

	logData["state"] = state
	log.Info("about to check resources exist", logData)

	if err := api.dataStore.Backend.CheckDatasetExists(id, state); err != nil {
		log.ErrorC("unable to find dataset", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getEditionsAction, audit.Unsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return

		}
		handleErrorType(editionDocType, err, w)
		return
	}

	results, err := api.dataStore.Backend.GetEditions(id, state)
	if err != nil {
		log.ErrorC("unable to find editions for dataset", err, logData)

		if auditErr := api.auditor.Record(r.Context(), getEditionsAction, audit.Unsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}

		handleErrorType(editionDocType, err, w)
		return
	}

	var logMessage string
	var b []byte

	if authorised {

		// User has valid authentication to get raw edition document
		b, err = json.Marshal(results)
		if err != nil {
			log.ErrorC("failed to marshal a list of edition resources into bytes", err, logData)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logMessage = "get all editions with auth"

	} else {

		// User is not authenticated and hance has only access to current sub document
		var publicResults []*models.Edition
		for i := range results.Items {
			publicResults = append(publicResults, results.Items[i].Current)
		}

		b, err = json.Marshal(&models.EditionResults{Items: publicResults})
		if err != nil {
			log.ErrorC("failed to marshal a list of public edition resources into bytes", err, logData)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logMessage = "get all editions without auth"
	}

	if auditErr := api.auditor.Record(r.Context(), getEditionsAction, audit.Successful, auditParams); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		log.Error(err, logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Debug(logMessage, log.Data{"dataset_id": id})
}

func (api *DatasetAPI) getEdition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	editionID := vars["edition"]
	logData := log.Data{"dataset_id": id, "edition": editionID}
	auditParams := common.Params{"dataset_id": id, "edition": editionID}

	if auditErr := api.auditor.Record(r.Context(), getEditionAction, audit.Attempted, auditParams); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	authorised, logData := api.authenticate(r, logData)

	var state string
	if !authorised {
		state = models.PublishedState
	}

	if err := api.dataStore.Backend.CheckDatasetExists(id, state); err != nil {
		log.ErrorC("unable to find dataset", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getEditionAction, audit.Unsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}
		handleErrorType(editionDocType, err, w)
		return
	}

	edition, err := api.dataStore.Backend.GetEdition(id, editionID, state)
	if err != nil {
		log.ErrorC("unable to find edition", err, logData)
		if auditErr := api.auditor.Record(r.Context(), getEditionAction, audit.Unsuccessful, auditParams); auditErr != nil {
			handleAuditingFailure(w, auditErr, logData)
			return
		}
		handleErrorType(editionDocType, err, w)
		return
	}

	var logMessage string
	var b []byte

	if authorised {

		// User has valid authentication to get raw edition document
		b, err = json.Marshal(edition)
		if err != nil {
			log.ErrorC("failed to marshal edition resource into bytes", err, logData)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logMessage = "get edition with auth"

	} else {

		// User is not authenticated and hance has only access to current sub document
		b, err = json.Marshal(edition.Current)
		if err != nil {
			log.ErrorC("failed to marshal public edition resource into bytes", err, logData)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logMessage = "get public edition without auth"
	}

	if auditErr := api.auditor.Record(r.Context(), getEditionAction, audit.Successful, auditParams); auditErr != nil {
		handleAuditingFailure(w, auditErr, logData)
		return
	}

	setJSONContentType(w)
	_, err = w.Write(b)
	if err != nil {
		log.Error(err, logData)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Debug(logMessage, logData)
}
