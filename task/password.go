package task

import (
	"crypto/sha256"
	"fmt"
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
   wrapPasswordReader assigns the password returned by the passed
   PasswordReader to the corresponding variables and then
   returns them.
*/
func wrapPasswordReader(pr PasswordReader) ([]byte, error) {
	password, err := pr.ReadPassword()
	if err != nil {
		return nil, err
	}
	return password, nil
}

/*
   ReadPassword receives the inputed user password from wrapPasswordReader.
   It accepts a bool var that informs the function if this is a new
   database. If true then the user is prompted to confirm the previously
   given password. If the confirmation fails an error is returned.

   Implementation details:
   - wrapPasswordReader will call whichever ReadPassword method was passed
     to it altering how the password is fetched. However, the validity
     checks of the given password should be contained in this function.
*/
func ReadPassword(newDb bool, pr PasswordReader) (password []byte, err error) {
	fmt.Printf("Enter Password: ")
	password, err = wrapPasswordReader(pr)
	if err != nil {
		return nil, err
	}

	if newDb == true {
		fmt.Printf("\nConfirm Password: ")
		confirmPass, err := wrapPasswordReader(pr)
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
   HashPassword returns a byte slice of the sha256 hash created from
   the password inputed by the user.

   Implementation details:
   - It does not salt the password, it just returns the sha256 equivalent hash.
*/
func HashPassword(password []byte) []byte {
	initHash := sha256.Sum256(password)
	hash := initHash[:]

	return hash
}

/*
   regPassword calls HashPassword and writes the returned key into a a file.
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

	key := HashPassword(password)

	err = os.WriteFile(path.KEY, key, 0600)
	if err != nil {
		log.Fatal(err)
	}
}
