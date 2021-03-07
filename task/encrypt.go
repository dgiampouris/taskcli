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

/*
   readPassword reads the inputed user password from the cli (stdin).
   It accepts a bool var that informs the function if this is a new
   database. If true then the user is prompted to confirm the previously
   given password. If the confirmation fails, execution ends fatally.

   Implementation details:
   - The term package is used, specifically the fucntion term.ReadPassword
     which temporarily changes the prompt and reads a password without echo.
*/
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

/*
   hashPassword returns a byte slice of the sha256 hash created from
   the password inputed by the user.

   Implementation details:
   - It does not salt the password, it just returns the sha256 equivalent hash.
*/
func hashPassword(password []byte) []byte {
	initHash := sha256.Sum256(password)
	hash := initHash[:]

	return hash
}

/*
   regPassword calls hashPassword and writes the returned key into a a file.
   It uses SetPaths to get the path where the key is stored.

   Implementation details:
   - Permissions are set so that only the user can read/write the key file.
*/
func regPassword(newDb bool) {
	var path Path = *SetPaths()
	key := hashPassword(readPassword(newDb))
	err := os.WriteFile(path.key, key, 0600)
	if err != nil {
		log.Panic(err)
	}
}

/*
   dbEncrypt encrypts the file where the boltdb database exists.
   It initialy loads the db into the memory as well as the key.
   It then creates a new cipher block with a given key of 32 bytes (AES-256).
   Based on the initial cipher block, a block cipher wrapped in
   Galois Counter Mode with the standard nonce length is returned.
   After a nonce is created with the proper length, the data is
   encrypted by calling Seal and written back to the filesystem
   at the path provided by path.db.

   Implementation details:
   - Should be called with defer in each function that interacts with the db.
*/
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

	// gcm.NonceSize returns the size of the nonce that must be
	// passed to Seal.
	// Never use more than 2^32 random nonces with a given key.
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
   - There are 3 total attempts to decrypt the file.
   - A boolean indicating whether the decryption was successful
     or not is returned.
*/
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

		// The nonce value is stored at the beginning of the
		// encrypted file.
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
