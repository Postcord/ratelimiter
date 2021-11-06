package ratelimiter

type ErrUnavailable struct {
	msg string
}

func (e *ErrUnavailable) Error() string {
	return e.msg
}

func NewErrUnavailable(msg string) *ErrUnavailable {
	return &ErrUnavailable{msg: msg}
}
