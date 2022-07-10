package utils

import (
	"fmt"
	"github.com/isd-sgcu/rnkm65-file/src/constant"
	"strings"
	"time"
)

func GetObjectName(filename string, secret string, fileType constant.FileType) (string, error) {
	text := fmt.Sprintf("%s%s%v", filename, secret, time.Now().Unix())
	hashed := Hash([]byte(text))

	hashed = strings.ReplaceAll(hashed, "/", "")

	switch fileType {
	case constant.FILE:
		return fmt.Sprintf("file-%s-%d-%s", filename, time.Now().Unix(), hashed), nil
	case constant.IMAGE:
		return fmt.Sprintf("image-%s-%d-%s", filename, time.Now().Unix(), hashed), nil
	default:
		return "", nil
	}
}
