package domain

type ReqFormData struct {
	FileBytes []byte `json:"file_bytes"`
	FileName  string `json:"file_name"`
}

func (req ReqFormData) GetFileBytes() []byte {
	return req.FileBytes
}
func (req ReqFormData) GeFileName() string {
	return req.FileName
}

type ReqFormDataImp interface {
	GetFileBytes() []byte
	GeFileName() string
}
