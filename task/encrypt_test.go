package task

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
   TestEncryptDecrypt test if the encryption and decryption functions
   have the desired effect to the file that is being encrypted or
   decrypted. Initially a mock password and data are defined. The
   original key and db is backed up and then the mock data are
   encrypted and decrypted with the mock key. If the resulting
   data is the same after the whole process, then the encryption
   process is successful.

   Implementation details:
   - It is also checked if the encrypted data is different than
     the original data, to make sure that the encryption indeed
     took place.
*/
func TestEncryptDecrypt(t *testing.T) {
	var path Path = *SetPaths()
	password := []byte("password")
	originData := []byte("data")

	_ = os.Rename(path.DB, path.DB+".bak")
	_ = os.Rename(path.KEY, path.KEY+".bak")

	_ = os.Remove(path.DB)
	_ = os.Remove(path.KEY)

	key := HashPassword(password)
	_ = os.WriteFile(path.KEY, key, 0600)
	_ = os.WriteFile(path.DB, originData, 0644)

	DbEncrypt()
	encData, _ := os.ReadFile(path.DB)
	assert.NotEqual(t, originData, encData)

	DbDecrypt()
	decData, _ := os.ReadFile(path.DB)
	assert.Equal(t, originData, decData)

	os.Remove(path.DB)
	os.Remove(path.KEY)

	_ = os.Rename(path.DB+".bak", path.DB)
	_ = os.Rename(path.KEY+".bak", path.KEY)
}
