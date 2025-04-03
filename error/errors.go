package error

import "errors"

// BackendNotFoundError ingress中的svc在当前namespace不存在
type BackendNotFoundError struct {
	errMsg string
}

func (e BackendNotFoundError) Error() string {
	return e.errMsg
}

func IsBackendNotFoundError(e error) bool {
	var err BackendNotFoundError
	return errors.As(e, &err)
}

func NewBackendNotFoundError(errMsg string) error {
	return BackendNotFoundError{
		errMsg: errMsg,
	}
}

// CtrNotFoundError ingress中的svc的annotations或ingressClass有没有选择当前控制器
type CtrNotFoundError struct {
	errMsg string
}

func (e CtrNotFoundError) Error() string {
	return e.errMsg
}

func IsCtrNotFoundError(e error) bool {
	var err CtrNotFoundError
	return errors.As(e, &err)
}

func NewCtrNotFoundError(errMsg string) error {
	return CtrNotFoundError{
		errMsg: errMsg,
	}
}
