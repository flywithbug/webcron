package models


type Response map[string]interface{}

func NewResponse() Response {
	return make(Response)
}

func ErrorResponse(errno int, msg string) Response {
	r := make(Response)
	r.SetErrorInfo(errno, msg)
	return r
}

func (s Response) SetErrorInfo(errno int, msg string) {
	s["code"] = errno
	s["msg"] = msg
}
func (s Response) SetSuccessInfo(code int, msg string) {
	s["code"] = code
	s["msg"] = msg
}

func (s Response) AddResponseInfo(key string, val interface{}) {
	s[key] = val
}
