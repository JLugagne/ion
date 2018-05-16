package types

import (
	"encoding/json"
	"net/http"

	"github.com/lawrencegripper/ion/internal/pkg/common"
)

// cSpell:ignore bson

//MetadataProvider is a document storage DB for storing document data
type MetadataProvider interface {
	GetEventContextByID(id string) (*EventContext, error)
	CreateEventContext(metadata *EventContext) error
	CreateInsight(insight *Insight) error
	Close()
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	GetBlobs(outputDir string, filePaths []string) error
	PutBlobs(filePaths []string) (map[string]string, error)
	Close()
}

//EventPublisher is responsible for publishing events to a remote system
type EventPublisher interface {
	Publish(e common.Event) error
	Close()
}

//EventContext is a single entry in a document
type EventContext struct {
	*common.Context
	ParentEventID string               `bson:"parentEventId" json:"parentEventId"`
	Files         []string             `bson:"files" json:"files"`
	Data          common.KeyValuePairs `bson:"data" json:"data"`
}

//Insight is used to export structure data
type Insight struct {
	*common.Context
	ExecutionID string               `bson:"id" json:"id"`
	Data        common.KeyValuePairs `bson:"data" json:"data"`
}

//ErrorResponse is a struct intended as JSON HTTP response
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

//Send returns a structured error object
func (e *ErrorResponse) Send(w http.ResponseWriter) {
	w.Header().Set(ContentType, ContentTypeApplicationJSON)
	w.WriteHeader(e.StatusCode)
	_ = json.NewEncoder(w).Encode(e.Message)
}