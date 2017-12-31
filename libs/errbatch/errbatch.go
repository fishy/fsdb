package errbatch

import (
	"bytes"
	"fmt"
)

// ErrBatch is an error that can contain multiple errors.
type ErrBatch struct {
	errors []error
}

// NewErrBatch creates a new ErrBatch.
func NewErrBatch() *ErrBatch {
	return &ErrBatch{
		errors: make([]error, 0),
	}
}

// Error satisifies the error interface.
func (eb *ErrBatch) Error() string {
	var buf bytes.Buffer
	buf.WriteString(
		fmt.Sprintf("total %d error(s) in this batch", len(eb.errors)),
	)
	for i, err := range eb.errors {
		if i == 0 {
			buf.WriteString(": ")
		} else {
			buf.WriteString("; ")
		}
		buf.WriteString(err.Error())
	}
	return buf.String()
}

// Add addes an error into the batch.
//
// If the error is also an ErrBatch,
// its underlying errors will be added instad of the ErrBatch itself.
//
// Any nil error(s) will be skipped.
func (eb *ErrBatch) Add(err error) {
	if batch, ok := err.(*ErrBatch); ok {
		for _, e := range batch.errors {
			eb.Add(e)
		}
		return
	}
	if err != nil {
		eb.errors = append(eb.errors, err)
	}
}

// Compile compiles the batch.
//
// If the batch contains zero errors, it will return nil.
//
// If the batch contains exactly one error,
// that underlying error will be returned.
//
// Otherwise, the batch itself will be returned.
func (eb *ErrBatch) Compile() error {
	switch len(eb.errors) {
	case 0:
		return nil
	case 1:
		return eb.errors[0]
	default:
		return eb
	}
}

// Clear clears the batch.
func (eb *ErrBatch) Clear() {
	eb.errors = make([]error, 0)
}

// GetErrors returns a copy of the underlying errors
func (eb *ErrBatch) GetErrors() []error {
	errors := make([]error, len(eb.errors))
	for i, err := range eb.errors {
		errors[i] = err
	}
	return errors
}
