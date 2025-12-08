package domain

type ReqParamsImp interface {
	H(rd RespData)
}

type RespData struct {
	Msg    string `json:"msg"`
	Code   int    `json:"code"`
	Status int    `json:"status"`
}
