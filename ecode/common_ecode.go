package ecode

var (
	OK                 = add(0, "正确")
	AppKeyInvalid      = add(-1, "应用程序不存在或已被封禁")
	AccessKeyErr       = add(-2, "Access Key错误")
	SignCheckErr       = add(-3, "API 校验密匙错误")
	RequestErr         = add(-400, "请求错误")
	Unauthorized       = add(-401, "未认证")
	AccessDenied       = add(-403, "访问权限不足")
	NothingFound       = add(-404, "啥都木有")
	MethodNotAllowed   = add(-405, "不支持该方法")
	AccessTokenExpires = add(-406, "Token 过期")
	AccessTokenInvalid = add(-407, "Token 无效")
	ServerErr          = add(-500, "服务器错误")
	ServiceUnavailable = add(-503, "过载保护,服务暂不可用")
	Deadline           = add(-504, "服务调用超时")
	LimitExceed        = add(-509, "超出限制")

	// ws
	WSBufferFull          = add(-800, "websocket 缓冲区已满")
	WSClientNotExist      = add(-801, "websocket 客户端不存在")
	WSClientIdIsNull      = add(-802, "websocket 客户端ID为空")
	WSMessageHandlerIsNil = add(-803, "没有websocket 消息处理函数")
	// jwt
	SigningKeyIsNull = add(-804, "签名密钥为空")
	SigningKeyLimit  = add(-805, "签名密钥长度不足,至少32字节: 实际 %s 字节")

	// sse
	SSEBufferFull     = add(-806, "sse 缓冲区已满")
	SSEClientNotExist = add(-807, "sse 客户端不存在")
	SSEClientIdIsNull = add(-808, "sse 客户端ID为空")
)
