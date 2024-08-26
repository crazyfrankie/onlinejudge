package dao

import (
	"context"

	"gorm.io/gorm"

	"oj/problem/domain"
)

type ProblemDao interface {
	getDifficulty(diff uint8) string
	setDifficulty(diff string) uint8
	Create(ctx context.Context, problem domain.Problem) error
	Update(ctx context.Context, problem domain.Problem) (domain.Problem, error)
	Delete(ctx context.Context, id uint64) error
	FindById(ctx context.Context, id uint64) (domain.Problem, error)
}

type GormProblemDao struct {
	db *gorm.DB
}

func (dao *GormProblemDao) getDifficulty(diff uint8) string {
	diffMap := map[uint8]string{
		0: "Easy",
		1: "Medium",
		2: "Hard",
	}
	diffString := diffMap[diff]
	return diffString
}

func (dao *GormProblemDao) setDifficulty(diff string) uint8 {
	diffMap := map[string]uint8{
		"Easy":   0,
		"Medium": 1,
		"Hard":   2,
	}
	dif := diffMap[diff]
	return dif
}

func (dao *GormProblemDao) Create(ctx context.Context, problem domain.Problem) error {

}

func (dao *GormProblemDao) Update(ctx context.Context, problem domain.Problem) (domain.Problem, error) {

}

func (dao *GormProblemDao) Delete(ctx context.Context, id uint64) error {

}

func (dao *GormProblemDao) FindByTag(ctx context.Context, tag string) ([]domain.Problem, error) {

}

func (dao *GormProblemDao) FindById(ctx context.Context, id uint64) (domain.Problem, error) {

}
