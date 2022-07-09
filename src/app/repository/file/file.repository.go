package file

import (
	"github.com/isd-sgcu/rnkm65-file/src/app/model/file"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByOwnerID(id string, result *file.File) error {
	return r.db.First(&result, "owner_id = ?", id).Error
}

func (r *Repository) CreateOrUpdate(result *file.File) error {
	if r.db.Where("owner_id = ?", result.OwnerID).Updates(&result).RowsAffected == 0 {
		return r.db.Create(&result).Error
	}
	return nil
}

func (r *Repository) Delete(id string) error {
	return r.db.First(id).Delete(&file.File{}).Error
}
