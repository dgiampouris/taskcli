package task

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"strconv"
)

type Path struct {
	db  string
	key string
}

func SetPaths() *Path {
	db := os.Getenv("HOME") + "/.tasks.db"
	key := "/dev/shm/.taskdb"
	return &Path{db: db, key: key}
}

/*
   dbOpen opens an existing boltdb file or creates a new one
   if it doesn't already exist. The file's name is
   tasks.db, the permissions allow the owner to
   read and write, the group to read only and others
   to read only. It returns a pointer to a DB type.
*/
func dbOpen() *bolt.DB {
	var path Path = *SetPaths()
	if _, err := os.Stat(path.db); err == nil {
		success := dbDecrypt()
		if success == false {
			log.Fatalf("Decryption failed.\n")
		}
	} else if os.IsNotExist(err) {
		newDb := true
		regPassword(newDb)
	}

	db, err := bolt.Open(path.db, 0644, nil)
	if err != nil {
		log.Panic(err)
	}

	return db
}

/*
   AddTask accepts a string and stores it into an existing boltdb
   database. All the tasks are contained into a bucket named "Tasks".
   Since boltdb is a k/v store, the key of each task is a number, and
   the value is the task itself. The key and the value is stored as
   a slice of bytes. Additionally, there is one key named "lastIndex"
   which tracks the number of all the tasks ever added in the database.
   lastIndex is used as the key for each task added in the database.
   Everytime a task is added, lastIndex is incremented by 1.

   Implementation details:
   - strconv.Itoa is used to convert int->string, which is then converted
     to []byte, since boltdb stores everything as []byte.
   - If lastIndex doesn't exist in the bucket then it gets created and
     given the value of 0.
   - If the value of lastIndex goes over the maximum value for a 32 bit
     integer then lastIndex is set back to 0, in order to avoid an integer
     overflow.
   - strconv.Atoi and string() are used to convert lastIndex's value to an integer.
*/
func AddTask(task string) {
	db := dbOpen()
	defer dbEncrypt()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		/*
			// Set initial index value and prevent integer overflow
			tmp := bucket.Get([]byte("lastIndex"))
			if tmp == nil {
				bucket.Put([]byte("lastIndex"), []byte(strconv.Itoa(0)))
			} else if string(tmp) == string(math.MaxInt32) {
				bucket.Put([]byte("lastIndex"), []byte(strconv.Itoa(0)))
			}

			lastIndex, err := strconv.Atoi(string(bucket.Get([]byte("lastIndex"))))
			if err != nil {
				return fmt.Errorf("strconv atoi: %s", err)
			}
			lastIndex++
		*/

		// Generate an index for the tasks
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so the error check is ignored.

		index, _ := bucket.NextSequence()
		id := strconv.Itoa(int(index))

		err = bucket.Put([]byte(id), []byte(task))
		if err != nil {
			return fmt.Errorf("bucket put new task: %s", err)
		}

		/*
			err = bucket.Put([]byte("lastIndex"), []byte(strconv.Itoa(lastIndex)))
			if err != nil {
				return fmt.Errorf("bucket put lastIndex: %s", err)
			}
		*/

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

/*
   ListTasks iterates over all the keys and values within the
   boltdb database and prints each key and each value.
   If the key == "lastIndex" then printing that key and value
   is skipped since there is no reason for the user to know
   the number of the last index.

   Implementation details:
   - Returning nil in the ForEach method is like a loop's
     continue statement. If err is returned then the
     iteration is stopped and the error is returned to the caller.
*/
func ListTasks() {
	db := dbOpen()
	defer dbEncrypt()
	defer db.Close()

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		fmt.Printf("\nHere's a list of your tasks:\n\n")

		err = bucket.ForEach(func(key, val []byte) error {
			//if string(key) == "lastIndex" {
			//		return nil
			//	}
			fmt.Printf("%s. %s\n", key, val)
			return nil
		})

		if err != nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}

/*
   DeleteTask deletes a task based on the index number provided.
   The only bucket to ever exist in the database is "Tasks" so
   the deletion only takes place in that bucket.

   Implementation details:
   - Deleting lastIndex is prohibited. No error is printed in
     case of an invalid index number.
*/
func DeleteTask(taskNum string) {
	db := dbOpen()
	defer dbEncrypt()
	defer db.Close()

	//if taskNum != "lastIndex" {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("Tasks"))
		err := tx.Bucket([]byte("Tasks")).Delete([]byte(taskNum))

		cursor := bucket.Cursor()

		for key, val := cursor.Seek([]byte(taskNum)); key != nil; key, val = cursor.Next() {
			temp, _ := strconv.Atoi(string(key))
			fmt.Println(temp, string(val))

			nextKey := []byte(strconv.Itoa(temp))
			nextVal := val

			currentKey := []byte(strconv.Itoa(temp - 1))
			err = bucket.Put(currentKey, nextVal)
			if err != nil {
				return fmt.Errorf("bucket re-order task: %s", err)
			}

			err = bucket.Delete(nextKey)
			if err != nil {
				return fmt.Errorf("bucket re-order task: %s", err)
			}

			afterNextKey := []byte(strconv.Itoa(temp + 1))
			afterNextVal := bucket.Get(afterNextKey)
			if afterNextVal == nil {
				err = bucket.SetSequence(uint64(temp - 1))
				if err != nil {
					return fmt.Errorf("bucket re-order task: %s", err)
				}

				break
			}

		}
		return err
	})

	if err != nil {
		log.Panic(err)
	}
	//}
}
