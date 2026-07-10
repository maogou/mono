package errno

var (
	Ok                  = &ErrNo{Code: 0, Message: "OK"}
	InternalServerError = &ErrNo{Code: 500, Message: "服务器内部,请稍后再试!"}
)
