package app

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/lawrencegripper/ion/dispatcher/messaging"
	"github.com/twinj/uuid"
	"os"
)

//CompareHash compares a secret string against a hash
func CompareHash(secret, secretHash string) error {
	if secret == "" {
		return fmt.Errorf("secret header was not found")
	}
	if Hash(secret) != secretHash {
		return fmt.Errorf("secret did not match")
	}
	return nil
}

//Hash returns a MD5 hash of the provided string
func Hash(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s)) // nolint: errcheck
	return hex.EncodeToString(hasher.Sum(nil))
}

//MustNotBeEmpty panics if any of the strings provided are empty
func MustNotBeEmpty(strs ...string) {
	for _, s := range strs {
		if s == "" {
			panic("required string is empty")
		}
	}
}

//MustNotBeNil panics if any of the objects provided are nil
func MustNotBeNil(objs ...interface{}) {
	for _, o := range objs {
		if o == nil {
			panic("required obj is nil")
		}
	}
}

//ClearDir removes all the content from a directory
func ClearDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed removing directory path '%s' with error: '%+v'", dirPath, err)
	}
	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("failed creating directory path '%s' with error: '%+v'", dirPath, err)
	}
	return nil
}

//RemoveFile removes a file from the file system
func RemoveFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to remove file at path '%s' with error: '%+v'", filePath, err)
	}
	return nil
}

//NewGUID generates a new guid as a string
func NewGUID() string {
	guid := fmt.Sprintf("%v", uuid.NewV4())
	return guid
}

//Remove removes an entry from a key value pair array
func Remove(s []messaging.KeyValuePair, i int) []messaging.KeyValuePair {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
