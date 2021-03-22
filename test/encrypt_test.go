package tasktest

import (
	"os"
	"testing"

	"github.com/dgiampouris/taskcli/task"
	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	var path task.Path = *task.SetPaths()
	password := []byte("password")
	originData := []byte("data")

	_ = os.Rename(path.DB, path.DB+".bak")
	_ = os.Rename(path.KEY, path.KEY+".bak")

	_ = os.Remove(path.DB)
	_ = os.Remove(path.KEY)

	key := task.HashPassword(password)
	_ = os.WriteFile(path.KEY, key, 0600)
	_ = os.WriteFile(path.DB, originData, 0644)

	task.DbEncrypt()
	encData, _ := os.ReadFile(path.DB)
	assert.NotEqual(t, originData, encData)

	task.DbDecrypt()
	decData, _ := os.ReadFile(path.DB)
	assert.Equal(t, originData, decData)

	os.Remove(path.DB)
	os.Remove(path.KEY)

	_ = os.Rename(path.DB+".bak", path.DB)
	_ = os.Rename(path.KEY+".bak", path.KEY)
}
