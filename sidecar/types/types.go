package types

import (
	"encoding/json"
	"net/http"
)

//MetadataProvider is a document storage DB for storing document data
type MetadataProvider interface {
	GetEventContextByID(id string) (*Metadata, error)
	CreateEventContext(metadata *Metadata) error
	CreateInsight(insight *Insight) error
	Close()
}

//BlobProvider is responsible for getting information about blobs stored externally
type BlobProvider interface {
	GetBlobs(outputDir string, filePaths []string) error
	PutBlobs(filePaths []string) error
	Close()
}

//EventPublisher is responsible for publishing events to a remote system
type EventPublisher interface {
	Publish(e Event) error
	Close()
}

//TODO: Rename to eventContext
//Metadata is a single entry in a document
type Metadata struct {
	EventID       string         `bson:"id" json:"id"`
	CorrelationID string         `bson:"correlationId" json:"correlationId"`
	ParentEventID string         `bson:"parentEventId" json:"parentEventId"`
	Files         []string       `bson:"files" json:"files"`
	Data          []KeyValuePair `bson:"data" json:"data"`
}

//Insight todo
type Insight struct {
	ExecutionID   string         `bson:"id" json:"id"`
	CorrelationID string         `bson:"correlationId" json:"correlationId"`
	EventID       string         `bson:"eventId" json:"eventId"`
	ParentEventID string         `bson:"parentEventId" json:"parentEventId"`
	Data          []KeyValuePair `bson:"data" json:"data"`
}

//KeyValuePair is a key value pair
type KeyValuePair struct {
	Key   string      `bson:"key" json:"key"`
	Value interface{} `bson:"value" json:"value"`
}

//Event the basic event data format
type Event struct {
	EventID        string         `json:"eventID"`
	Type           string         `json:"type"`
	PreviousStages []string       `json:"previousStages"`
	CorrelationID  string         `json:"correlationID"`
	Data           []KeyValuePair `json:"data"`
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
