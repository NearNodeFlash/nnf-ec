package common

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"stash.us.cray.com/rabsw/ec"
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

	if err != nil {
		// If the supplied error is of an Element Controller Controller Error type,
		// extract the status code from the error to return as the response.
		var e *ec.ControllerError
		if errors.As(err, &e) {
			// TODO: Make some form of a redfish error response
			w.WriteHeader(e.StatusCode)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	if s != nil {
		response, err := json.Marshal(s)
		if err != nil {
			log.WithError(err).Error("Failed to marshal json response")
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = w.Write(response)
		if err != nil {
			log.WithError(err).Error("Failed to write json response")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

const (
	NamespaceMeatadataSignature = 0x54424252 // "RBBT"
	NamespaceMeatadataRevision  = 1
)

type NamespaceMetadata struct {
	Signature uint32
	Revision  uint16
	Rsvd      uint16
	Index     uint16
	Count     uint16
	Id        uuid.UUID
}

func EncodeNamespaceMetadata(pid uuid.UUID, index uint16, count uint16) ([]byte, error) {
	buf := new(bytes.Buffer)

	md := NamespaceMetadata{
		Signature: NamespaceMeatadataSignature,
		Revision:  NamespaceMeatadataRevision,
		Index:     index,
		Count:     count,
		Id:        pid,
	}

	err := binary.Write(buf, binary.LittleEndian, md)

	return buf.Bytes(), err
}

func DecodeNamespaceMetadata(buf []byte) (*NamespaceMetadata, error) {

	md := new(NamespaceMetadata)

	err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, md)

	return md, err
}
