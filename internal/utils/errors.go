package utils

import "os"

func IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
