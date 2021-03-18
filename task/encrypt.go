package task

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"golang.org/x/term"
)

// PasswordReader returns password read from a reader
type PasswordReader interface {
	ReadPassword() ([]byte, error)
}

// StdInPasswordReader represents an stdin password reader
type StdinPasswordReader struct {
}

/*
   ReadPassword for StdinPasswordReader reads the
   password from stdin by using the term package's
   ReadPassword function which prompts for user input.
*/
func (pr StdinPasswordReader) ReadPassword() ([]byte, error) {
	password, err := term.ReadPassword(syscall.Stdin)
	return password, err
}

/*
   initPassReader assigns the password returned by the passed
   PasswordReader to the corresponding variables and then
   returns them.
*/
func initPassReader(pr PasswordReader) ([]byte, error) {
	password, err := pr.ReadPassword()
	if err != nil {
		return nil, err
	}
	return password, nil
}

/*
   ReadPassword receives the inputed user password from initPassReader.
   It accepts a bool var that informs the function if this is a new
   database. If true then the user is prompted to confirm the previously
   given password. If the confirmation fails an error is returned.

   Implementation details:
   - initPassReader will call whichever ReadPassword method was passed
     to it altering how the password is fetched. However, the validity
     checks of the given password should be contained in this function.
*/
func ReadPassword(newDb bool, pr PasswordReader) (password []byte, err error) {
	fmt.Printf("Enter Password: ")
	password, err = initPassReader(pr)
	if err != nil {
		return nil, err
	}

	if newDb == true {
		fmt.Printf("\nConfirm Password: ")
		confirmPass, err := initPassReader(pr)
		if err != nil {
			return nil, err
		}

		if string(password) != string(confirmPass) {
			return nil, fmt.Errorf("\nPasswords did not match! Please try again.")

		}
	}

	return password, err
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
	pr := StdinPasswordReader{}

	password, err := ReadPassword(newDb, pr)
	if err != nil {
		log.Fatal(err)
	} else if password == nil {
		err = fmt.Errorf("\nPassword cannot be blank!\n")
		log.Fatal(err)
	}

	key := hashPassword(password)

	err = os.WriteFile(path.key, key, 0600)
	if err != nil {
		log.Fatal(err)
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
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Encryption Error: %v", p)
		}
	}()

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
   - TODO: Add details
*/
func dbDecrypt() {
	var path Path = *SetPaths()
	defer func() {
		if p := recover(); p != nil {
			os.Remove(path.key)
			fmt.Printf("Decryption Error: %v", p)
		}
	}()

	encryptedData, err := os.ReadFile(path.db)
	if err != nil {
		log.Panic(err)
	}

	decSuccess := true
	var data []byte

	key, err := os.ReadFile(path.key)
	if os.IsNotExist(err) {
		newDB := false
		regPassword(newDB)
		key, err = os.ReadFile(path.key)
	} else if err != nil {
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

	data, err = gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		decSuccess = false
	}

	if decSuccess == true {
		err = os.WriteFile(path.db, data, 0644)
		if err != nil {
			log.Panic(err)
		}
	} else if decSuccess == false {
		os.Remove(path.key)
		log.Fatal("\nDecryption error!\n")
	}
}
