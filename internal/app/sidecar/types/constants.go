package types

// cSpell:ignore inmemory, mongodb

const (
	//ContentType represents the 'Content-Type' HTTP header key
	ContentType string = "Content-Type"

	//ContentTypeApplicationJSON represents the 'Content-Type' HTTP header value "application/json"
	ContentTypeApplicationJSON string = "application/json"

	//MetaProviderMongoDB is a lowercase name for the MongoDB metadata provider
	MetaProviderMongoDB string = "mongodb"

	//MetaProviderInMemory is a lowercase name for the in-memory metadata provider
	MetaProviderInMemory string = "inmemory"

	//BlobProviderAzureStorage is a lowercase name for the azure blob storage blob provider
	BlobProviderAzureStorage string = "azureblob"

	//BlobProviderFileSystem is a lowercase name for the file system blob provider
	BlobProviderFileSystem string = "filesystem"

	//EventProviderServiceBus is a lowercase name for the file system blob provider
	EventProviderServiceBus string = "servicebus"

	//EventType is the key in the metadata to extract the event type using
	EventType string = "eventType"

	//FilesToInclude is the key in the metadata to extract the event type using
	FilesToInclude string = "files"

	//DevBaseDir is the base directory for development mode output
	DevBaseDir string = ".dev"
)