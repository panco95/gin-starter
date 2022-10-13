package resp

const (
	ERROR = 1
)

const (
	SERVER_ERROR        = "服务器错误"
	PARAM_INVALID       = "参数不合法"
	ACCOUNT_NOT_FOUND   = "账号不存在"
	ACCOUNT_PWD_ERROR   = "账号或密码错误"
	ACCOUNT_LOCKED      = "账号被封禁"
	CAPTCHA_ERROR       = "验证码错误"
	CAPTCHA_EXPIRED     = "验证码过期，请刷新后重试"
	TIMEOUT             = "登录超时"
	ACCOUNT_EXISTS      = "账号已存在"
	ACCOUNT_NOT_EXISTS  = "账号不存在"
	ACCOUNT_HAS_CHINESE = "用户名不能包含中文"
)

type Response struct {
	Code    int         `json:"code"`
	Result  interface{} `json:"result"`
	Message string      `json:"message"`
}

type PageResult struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}
