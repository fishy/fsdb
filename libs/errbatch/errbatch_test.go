package errbatch_test

import (
	"errors"
	"testing"

	"github.com/fishy/fsdb/libs/errbatch"
)

func TestAdd(t *testing.T) {
	err := errbatch.NewErrBatch()
	if len(err.GetErrors()) != 0 {
		t.Error("A new ErrBatch should contain zero errors.")
	}

	err.Add(nil)
	if len(err.GetErrors()) != 0 {
		t.Error("Nil errors should be skipped.")
	}

	err0 := errors.New("foo")
	err.Add(err0)
	if len(err.GetErrors()) != 1 {
		t.Error("Non-nil errors should be added to the batch.")
	}
	actual := err.GetErrors()[0]
	if actual != err0 {
		t.Errorf("Expected %#v, got %#v", err0, actual)
	}

	another := errbatch.NewErrBatch()
	err.Add(another)
	if len(err.GetErrors()) != 1 {
		t.Error("Empty batch should be skipped.")
	}
	err1 := errors.New("bar")
	another.Add(err1)
	err2 := errors.New("foobar")
	another.Add(err2)
	err.Add(another)
	if len(err.GetErrors()) != 3 {
		t.Error("The underlying errors should be added instead of the batch.")
	}

	batch := err.GetErrors()
	if batch[0] != err0 {
		t.Errorf("Expected %#v, got %#v", err0, batch[0])
	}
	if batch[1] != err1 {
		t.Errorf("Expected %#v, got %#v", err1, batch[1])
	}
	if batch[2] != err2 {
		t.Errorf("Expected %#v, got %#v", err2, batch[2])
	}

	err.Clear()
	if len(err.GetErrors()) != 0 {
		t.Error("A cleared ErrBatch should contain zero errors.")
	}
}

func TestCompile(t *testing.T) {
	batch := errbatch.NewErrBatch()
	err0 := errors.New("foo")
	err1 := errors.New("bar")
	err2 := errors.New("foobar")

	err := batch.Compile()
	if err != nil {
		t.Errorf("An empty batch should be compiled to nil, got: %#v", err)
	}
	batch.Add(err0)
	err = batch.Compile()
	if err != err0 {
		t.Errorf(
			"A single error batch should be compiled to %#v, got %#v",
			err0,
			err,
		)
	}
	batch.Add(err1)
	batch.Add(err2)
	err = batch.Compile()
	expect := "total 3 error(s) in this batch: foo; bar; foobar"
	if err.Error() != expect {
		t.Errorf("Compiled error expected %#v, got %v", expect, err)
	}
}
