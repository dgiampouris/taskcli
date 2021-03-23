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

// DataEncrypt contains the required methods for encryption and decryption
type DataEncrypt interface {
	Encrypt(data []byte, key []byte) ([]byte, error)
	Decrypt(data []byte, key []byte) ([]byte, error)
}

/*
   DB represents an abstraction over a db type. It's only used temporarily
   to create the needed encrypt and decrypt methods. In future version of
   the code this type should encompass the boltdb db type and extend it
   with encryption methods.

*/
type DB struct {
}

/*
   Encrypt is the method that encrypts the data passed as an argument and
   is ultimately responsible for encrypting the tasks database. This
   particular method creates a new cipher block with a given key of
   32 bytes (AES-256). Based on the initial cipher block, a block
   cipher wrapped in Galois Counter Mode with the standard nonce length
   is returned. After a nonce is created with the proper length,
   the data is encrypted by calling Seal. The encrypted data is then returned.
*/
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

/*
   Decrypt is the method that decrypts the data passed as an
   argument and is ultimately responsibel for decrypting the
   tasks database. Initially a new cipher block gets created
   with a given key of 32 bytes (AES-256). Based on the initial
   cipher block, a block cipher wrapped in Galois Counter Mode
   with the standard nonce length is returned. To decrypt, it is
   necessary to indicate the nonce value used during the encryption
   process. This value is saved at the beginning of the file.
   Open decrypts and authenticates the encrypted data. The data
   is then returned.
*/
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

/*
   wrapDataEncrypt is a wrapper aroung the Encrypt and Decrypt methods.
   It serves as a handler of the action that is to be performed which
   can be either encryption or decryption. In future iterations of the
   code this function can handle different kinds of encryption depending
   on what action or algorithm the developer will choose.
*/
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
   wrapDataEncrypt contains the desired actions for this function.
   After the encrypted data is returned it is written back to the filesystem
   at the path provided by path.DB.

   Implementation details:
   - DbEncrypt mainly acts as a second layer of wrapping around the encryption
     methods.
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
   It initialy loads the encrypted file into the memory as well as
   the key.if it does not exist then regPassword is called.
   wrapDataEncrypt contains the desired actions for the function.
   After the decrypted data is returned it is written back to the filesystem
   at the path provided by path.DB.

   Implementation details:
   - DbDecrypt mainly acts as a second layer of wrapping around the encryption
     methods, with some additional error handlind.
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
