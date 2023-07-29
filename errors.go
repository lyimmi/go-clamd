package clamd

import "errors"

// Exported errors
var (
	ErrDial               = errors.New("error while connecting to clamd")
	ErrCommandCall        = errors.New("error while calling clamd")
	ErrCommandRead        = errors.New("error while reading response from clamd")
	ErrEmptySrc           = errors.New("scan source is empty")
	ErrInvalidResponse    = errors.New("invalid response from clamd")
	ErrNoSuchFileOrDir    = errors.New("clamd can't find file or directory")
	ErrPermissionDenied   = errors.New("clamd can't open file or dir, permission denied")
	ErrSreamLimitExceeded = errors.New("clamd's INSTREAM size limit exceeded")
	ErrUnknown            = errors.New("unknown error")
	ErrEICARFound         = errors.New("Win.Test.EICAR_HDB-1 FOUND")
)
