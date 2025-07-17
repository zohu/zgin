package zgin

/**
 * 错误码预留：1-599
 * 1-99系统，100-599 HTTP CODE，600-999 框架预占
 */

// 信息响应 (100–199)
// 成功响应 (1,200–299)
// 重定向消息 (300–399)
// 客户端错误响应 (400–499)
// 服务端错误响应 (500–599)
const (
	MessageSuccess              MessageID = "1:ok"
	MessageParamInvalid         MessageID = "400:MessageParamInvalid"
	MessageLoginUnsupportedMode MessageID = "401:MessageLoginUnsupportedMode"
	MessageLoginFailed          MessageID = "401:MessageLoginFailed"
	MessageLoginTimeout         MessageID = "401:MessageLoginTimeout"
	MessageLoginIDUsed          MessageID = "401:MessageLoginIDUsed"
	MessageLoginTokenInvalid    MessageID = "401:MessageLoginTokenInvalid"
	MessageLoginSessionInvalid  MessageID = "401:MessageLoginSessionInvalid"
	MessageActionInvalid        MessageID = "403:MessageActionInvalid"
	MessagePathInvalid          MessageID = "404:MessagePathInvalid"
	MessageMethodInvalid        MessageID = "405:MessageMethodInvalid"
	MessageRequestInvalid       MessageID = "500:MessageRequestInvalid"
	MessageNotImplemented       MessageID = "501:MessageNotImplemented"
	MessageTimeout              MessageID = "504:MessageTimeout"
	MessageCreateFailed         MessageID = "600:MessageCreateFailed"
	MessageUpdateFailed         MessageID = "601:MessageUpdateFailed"
	MessageSaveFailed           MessageID = "602:MessageSaveFailed"
	MessageDeleteFailed         MessageID = "603:MessageDeleteFailed"
	MessageQueryFailed          MessageID = "604:MessageQueryFailed"
)
