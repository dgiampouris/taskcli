package task

import (
	"encoding/binary"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"log"
	"math"
	"os"
	"strconv"
)

type Path struct {
	db  string
	key string
}

/*
   SetPaths sets the default paths to store the db
   and the secret key. It serves as an initializer
   which is called in every function that needs to
   interact with the database or the key.

   Implementation details:
   - /dev/shm/ is mostly available on linux and not available in macOs
*/
func SetPaths() (path *Path) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Path Error: %v\n", p)
			path = &Path{db: "~/.tasks.db", key: "~/.tasksdbkey"}
		}
	}()

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Panic("$HOME environment variable not found! DB will be in ~/.\n")
	}

	db := os.Getenv("HOME") + "/.tasks.db"

	if stat, err := os.Stat("dev/shm/"); err == nil && !stat.IsDir() {
		log.Panic("/dev/shm does not exist, key file will be in ~/.\n")
	}

	key := "/dev/shm/.taskdb"
	return &Path{db: db, key: key}
}

/*
   itob returns an 8-byte big endian representation of val.
   Since everything is stored/retrieved as a []byte type
   from boltdb, and keys are byte-sorted, indexes
   need to be converted to the aforementioned representation.
*/
func itob(val int) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(val))
	return bytes
}

/*
   dbOpen opens an existing boltdb file or creates a new one
   if it doesn't already exist. The file's name is
   tasks.db, the permissions allow the owner to
   read and write, the group to read only and others
   to read only. It returns a pointer to a DB type.

   Implementation details:
   - TODO: talk about error handling
*/
func dbOpen() (db *bolt.DB) {
	var path Path = *SetPaths()
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Opening the database failed!\n")
			db = &bolt.DB{
				StrictMode:   false,
				NoSync:       false,
				NoGrowSync:   false,
				MmapFlags:    0,
				MaxBatchSize: 0,
				AllocSize:    0}
		}
	}()

	if _, err := os.Stat(path.db); err == nil {
		data := dbDecrypt()
		if data == nil {
			log.Fatal("Decryption error!")
		}

		err := os.WriteFile(path.db, data, 0644)
		if err != nil {
			log.Panic(err)
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
   a slice of bytes. Additionally, the NextSequence() method is used
   which returns an autoincrementing int to track the number of all
   the tasks ever added in the database. That int servers as an index
   which is used as the key for each task added in the database.

   Implementation details:
   - itob is used to convert the id into a byte representation.
   - The id has to be smaller than math.MaxInt32 to avoid int overflow.
   - The given task cannot be emtpy.
*/
func AddTask(task string) {
	db := dbOpen()
	defer dbEncrypt()
	defer db.Close()
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Adding the task failed!\n")
		}
	}()

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return err
		}

		// Generate an index for the tasks
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so the error check is ignored.
		index, _ := bucket.NextSequence()
		id := int(index)

		// Ensure id does not cause integer overflow
		if id < math.MaxInt32 && task != "" {
			err = bucket.Put(itob(id), []byte(task))
			if err != nil {
				return err
			}
		} else if id >= math.MaxInt32 {
			return fmt.Errorf("\nToo many tasks!\n")
		}

		return err
	})

	if err != nil {
		log.Panic(err)
	}
}

/*
   ListTasks iterates over all the keys and values within the
   boltdb database and prints each key and each value.
   The key is converted from []byte to int in order to be printed.

   Implementation details:
   - Returning nil in the ForEach method is like a loop's
     continue statement. If err is returned then the
     iteration is stopped and the error is returned to the caller.
*/
func ListTasks() {
	db := dbOpen()
	defer dbEncrypt()
	defer db.Close()
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Listing tasks failed!\n")
		}
	}()

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("Tasks"))
		if err != nil {
			return err
		}

		fmt.Printf("\nHere's a list of your tasks:\n\n")

		err = bucket.ForEach(func(key, val []byte) error {
			intKey := int(binary.BigEndian.Uint64(key))
			fmt.Printf("%d. %s\n", intKey, val)
			return nil
		})

		return err
	})

	if err != nil {
		log.Panic(err)
	}
}

/*
   DeleteTask deletes a task based on the index number provided.
   The only bucket to ever exist in the database is "Tasks" so
   the deletion only takes place in that bucket. When a task
   gets deleted, the keys of the tasks following the deleted
   task get decremented by 1.

   Implementation details:
   - If the id number is "" or < 1 then the deletion is skipped.
   - nextKey and nextVal are better names for key and val since
     cursor.Next() is executed before the first loop takes place.
*/
func DeleteTask(taskNum string) {
	db := dbOpen()
	defer dbEncrypt()
	defer db.Close()
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("Deleting the task failed!\n")
		}
	}()

	if taskNum != "" {
		err := db.Update(func(tx *bolt.Tx) error {
			id, err := strconv.Atoi(taskNum)
			if err != nil {
				return err
			} else if id < 1 {
				return fmt.Errorf("\nTasks can only have a positive non-zero id!\n")
			}

			taskIndex := itob(id)
			bucket := tx.Bucket([]byte("Tasks"))
			err = bucket.Delete(taskIndex)

			cursor := bucket.Cursor()
			for key, val := cursor.Seek(taskIndex); key != nil; key, val = cursor.Next() {
				nextKey := key
				nextVal := val

				currentKey := int(binary.BigEndian.Uint64(key)) - 1
				err = bucket.Put(itob(currentKey), nextVal)
				if err != nil {
					return err
				}

				err = bucket.Delete(nextKey)
				if err != nil {
					return err
				}
			}

			err = bucket.SetSequence(bucket.Sequence() - 1)
			if err != nil {
				return fmt.Errorf("\nSetting the bucket sequence failed!\n")
			}

			return err
		})

		if err != nil {
			log.Panic(err)
		}
	}
}
