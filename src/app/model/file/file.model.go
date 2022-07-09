package file

import (
	"github.com/isd-sgcu/rnkm65-file/src/app/model"
)

type File struct {
	model.Base
	Filename string `json:"filename" gorm:"index"`
	OwnerID  string `json:"owner_id" gorm:"index:,unique"`
	Tag      int    `json:"tag"`
}
