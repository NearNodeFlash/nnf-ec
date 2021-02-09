package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	ec "stash.us.cray.com/rabsw/nnf-ec/ec"
)

// Params -
func Params(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// UnmarshalRequest -
func UnmarshalRequest(r *http.Request, model interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &model)
}

// EncodeResponse -
func EncodeResponse(s interface{}, err error, w http.ResponseWriter) {
	if err != nil {
		log.WithError(err).Warn("Element Controller Error")
	}

	var e *ec.ControllerError
	if errors.As(err, &e) {
		w.WriteHeader(e.StatusCode)
		return
	}

	if s != nil {
		response, err := json.Marshal(s)
		if err != nil {
			log.WithError(err).Error("Failed to marshal json response")
			// TODO
		}
		_, err = w.Write(response)
		if err != nil {
			log.WithError(err).Error("Failed to write json response")
			// TODO
		}
	}
}
