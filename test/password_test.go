package tasktest

import (
	"errors"
	"testing"

	"github.com/dgiampouris/taskcli/task"
	"github.com/stretchr/testify/assert"
)

// stubPasswordReader is a mock implementation of StdinPasswordReader
type stubPasswordReader struct {
	Password    []byte
	ReturnError bool
}

/*
   ReadPassword is a mock implementation of the equivalent
   StdinPasswordReader method.
*/
func (pr stubPasswordReader) ReadPassword() ([]byte, error) {
	if pr.ReturnError {
		return nil, errors.New("stubbed error")
	}
	return pr.Password, nil
}

// TestReadPasswordReturnError tests if ReadPassword returns an error
func TestReadPasswordReturnError(t *testing.T) {
	newDb := false
	pr := stubPasswordReader{ReturnError: true}
	result, err := task.ReadPassword(newDb, pr)
	assert.Error(t, err)
	assert.Equal(t, errors.New("stubbed error"), err)
	assert.Equal(t, []byte(nil), result)
}

/*
   TestReadPassword tests if the password returned from ReadPassword
   is the same as the inputed one.
*/
func TestReadPassword(t *testing.T) {
	newDb := false
	pr := stubPasswordReader{Password: []byte("password")}
	result, err := task.ReadPassword(newDb, pr)
	assert.NoError(t, err)
	assert.Equal(t, []byte("password"), result)
}
