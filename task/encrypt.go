package task

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/term"
	"io"
	"log"
	"os"
	"syscall"
)

func readPassword(newDb bool) []byte {
	var confirmPass []byte

	fmt.Printf("Enter Password: ")
	password, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	if newDb == true {
		for i := 0; i < 3; i++ {
			fmt.Printf("\nConfirm Password: ")
			confirmPass, err = term.ReadPassword(syscall.Stdin)
			if err != nil {
				log.Fatal(err)
			}

			if string(password) == string(confirmPass) {
				break
			}
		}

		if string(password) != string(confirmPass) {
			log.Fatal("Passwords did not match! Please try again.")

		}
	}

	return password
}

func hashPassword(password []byte) []byte {
	initHash := sha256.Sum256(password)
	hash := initHash[:]

	return hash
}

func regPassword(newDb bool) {
	var path Path = *SetPaths()
	key := hashPassword(readPassword(newDb))
	err := os.WriteFile(path.key, key, 0600)
	if err != nil {
		log.Panic(err)
	}
}

// Should be called with defer in each function that interacts with the db
func dbEncrypt() {
	var path Path = *SetPaths()
	data, err := os.ReadFile(path.db)
	if err != nil {
		log.Panic(err)
	}

	key, err := os.ReadFile(path.key)
	if err != nil {
		log.Panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Panic(err)
	}

	// Never use more than 2^32 random nonces with a given key
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Panic(err)
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	err = os.WriteFile(path.db, encryptedData, 0644)
	if err != nil {
		log.Panic(err)
	}
}

func dbDecrypt() bool {
	var path Path = *SetPaths()
	decSuccess := true
	encryptedData, err := os.ReadFile(path.db)
	if err != nil {
		log.Panic(err)
	}

	for i := 0; i < 3; i++ {
		key, err := os.ReadFile(path.key)
		if os.IsNotExist(err) {
			newDB := false
			regPassword(newDB)
			key, err = os.ReadFile(path.key)
		} else if err != nil {
			log.Panic(err)
		}

		if err != nil {
			log.Panic(err)
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			log.Panic(err)
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			log.Panic(err)
		}

		nonce := encryptedData[:gcm.NonceSize()]
		encryptedData = encryptedData[gcm.NonceSize():]
		data, err := gcm.Open(nil, nonce, encryptedData, nil)
		if err != nil {
			decSuccess = false
			newDB := false
			log.Printf("Decryption error: %v\n", err)
			regPassword(newDB)
			continue
		}

		err = os.WriteFile(path.db, data, 0644)
		if err != nil {
			log.Panic(err)
		} else {
			decSuccess = true
			break
		}
	}

	return decSuccess
}
