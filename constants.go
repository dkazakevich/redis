package main

const (
	serverPort = "8080"

	baseUrl = "/api/v1/"

	itemNotFoundMsg = "Cache item not found"
	invalidIndexParamMsg = "Invalid `listIndex` param. Number required"
	valueNotListMsg = "Indicated value is not list"
	valueNotDictMsg = "Indicated value is not dictionary"
	dictItemNotFoundMsg = "Dictionary item not found"
	invalidPayloadRequestMsg = "Invalid payload request"
	invalidExpireValueMsg = "Invalid expire value"

	timeoutSetMsg = "The timeout was set"
	itemDeletedMsg = "Cache item deleted"
	dataPersistedMsg = "Cache Data persisted"
	dataReloadedMsg = "Cache Data reloaded"

	keyParam = "key"
	dictKeyParam = "dictKey"
	valueParam = "value"
	expireParam = "expire"
	listIndexParam = "listIndex"
	errorParam = "error"
	messageParam = "message"
)