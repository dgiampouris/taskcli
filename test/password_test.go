package tasktest

import (
	"errors"
	"testing"

	"github.com/dgiampouris/taskcli/task"
	"github.com/stretchr/testify/assert"
)

type stubPasswordReader struct {
	Password    []byte
	ReturnError bool
}

func (pr stubPasswordReader) ReadPassword() ([]byte, error) {
	if pr.ReturnError {
		return nil, errors.New("stubbed error")
	}
	return pr.Password, nil
}

func TestReadPasswordReturnError(t *testing.T) {
	newDb := false
	pr := stubPasswordReader{ReturnError: true}
	result, err := task.ReadPassword(newDb, pr)
	assert.Error(t, err)
	assert.Equal(t, errors.New("stubbed error"), err)
	assert.Equal(t, []byte(nil), result)
}

func TestReadPassword(t *testing.T) {
	newDb := false
	pr := stubPasswordReader{Password: []byte("password")}
	result, err := task.ReadPassword(newDb, pr)
	assert.NoError(t, err)
	assert.Equal(t, []byte("password"), result)
}
