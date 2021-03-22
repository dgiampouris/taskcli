package task

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
)

type DataEncrypt interface {
	Encrypt(data []byte, key []byte) ([]byte, error)
	Decrypt(data []byte, key []byte) ([]byte, error)
}

type DB struct {
}

func (db DB) Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// gcm.NonceSize returns the size of the nonce that must be
	// passed to Seal.
	// Never use more than 2^32 random nonces with a given key.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	return encryptedData, err
}

func (db DB) Decrypt(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// The nonce value is stored at the beginning of the
	// encrypted file.
	nonce := encryptedData[:gcm.NonceSize()]
	encryptedData = encryptedData[gcm.NonceSize():]

	data, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return data, err
}

func wrapDataEncrypt(db DataEncrypt, action string, data []byte, key []byte) ([]byte, error) {
	var err error

	if action == "encrypt" {
		data, err = db.Encrypt(data, key)
	} else if action == "decrypt" {
		data, err = db.Decrypt(data, key)
	} else {
		data = nil
		err = fmt.Errorf("\nUnknown encryption action choice!\n")
	}

	return data, err
}

/*
   dbEncrypt encrypts the file where the boltdb database exists.
   It initialy loads the db into the memory as well as the key.
   It then creates a new cipher block with a given key of 32 bytes (AES-256).
   Based on the initial cipher block, a block cipher wrapped in
   Galois Counter Mode with the standard nonce length is returned.
   After a nonce is created with the proper length, the data is
   encrypted by calling Seal and written back to the filesystem
   at the path provided by path.DB.

   Implementation details:
   - Should be called with defer in each function that interacts with the db.
*/
func DbEncrypt() {
	var path Path = *SetPaths()
	db := DB{}
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Encryption Error: %v", p)
		}
	}()

	data, err := os.ReadFile(path.DB)
	if err != nil {
		log.Panic(err)
	}

	key, err := os.ReadFile(path.KEY)
	if err != nil {
		log.Panic(err)
	}

	encryptedData, err := wrapDataEncrypt(db, "encrypt", data, key)
	if err != nil {
		log.Fatal(err)
	} else if encryptedData == nil {
		err = fmt.Errorf("\nEncrypted data is nil. Encryption failed!\n")
		log.Fatal(err)
	}

	err = os.WriteFile(path.DB, encryptedData, 0644)
	if err != nil {
		log.Panic(err)
	}
}

/*
   dbDecrypt decrypts the file where the boltdb database is located.
   It initialy loads the encrypted file into the memory.
   Then the key is loaded into the memory, but if it does not exist
   then regPassword is called. It then creates a new cipher block
   with a given key of 32 bytes (AES-256). Based on the initial
   cipher block, a block cipher wrapped in Galois Counter Mode
   with the standard nonce length is returned.
   To decrypt, it is necessary to indicate the nonce value used
   during the encryption process. This value is saved at the
   beginning of the file. Open decrypts and authenticates
   the encrypted data.

   Implementation details:
   - TODO: Add details
*/
func DbDecrypt() {
	var path Path = *SetPaths()
	db := DB{}
	defer func() {
		if p := recover(); p != nil {
			os.Remove(path.KEY)
			fmt.Printf("Decryption Error: %v", p)
		}
	}()

	encryptedData, err := os.ReadFile(path.DB)
	if err != nil {
		log.Panic(err)
	}

	key, err := os.ReadFile(path.KEY)
	if os.IsNotExist(err) {
		newDB := false
		regPassword(newDB)
		key, err = os.ReadFile(path.KEY)
	} else if err != nil {
		log.Panic(err)
	}

	data, err := wrapDataEncrypt(db, "decrypt", encryptedData, key)
	if err == nil {
		err = os.WriteFile(path.DB, data, 0644)
		if err != nil {
			log.Panic(err)
		}
	} else if err != nil {
		os.Remove(path.KEY)
		log.Fatal("\nDecryption error!\n")
	}
}
