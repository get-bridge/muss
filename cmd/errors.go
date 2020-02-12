package cmd

// QuietError will print its message without usage information.
type QuietError struct {
	err error
}

// NewQuietError takes a return and returns a new QuietError.
func NewQuietError(err error) *QuietError {
	return &QuietError{
		err: err,
	}
}

// Error returns the error message string.
func (e *QuietError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error.
func (e *QuietError) Unwrap() error {
	return e.err
}

// QuietErrorOrNil returns nil if passed nil otherwise wraps the provided error
// in a QuietError.
func QuietErrorOrNil(err error) error {
	if err == nil {
		return nil
	}
	return NewQuietError(err)
}
