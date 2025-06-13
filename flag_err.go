package zgin

const (
	MessageSuccess          MessageID = "1:ok"
	MessageInvalidParameter MessageID = "400:MessageInvalidParameter"
	MessageInvalidToken     MessageID = "401:MessageInvalidToken"
	MessageInvalidSession   MessageID = "401:MessageInvalidSession"
	MessageInvalidPath      MessageID = "404:MessageInvalidPath"
	MessageInvalidMethod    MessageID = "405:MessageInvalidMethod"
	MessageInvalidRequest   MessageID = "500:MessageInvalidRequest"
	MessageNotImplemented   MessageID = "501:MessageNotImplemented"
)
