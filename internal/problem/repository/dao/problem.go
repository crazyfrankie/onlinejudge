package dao

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // 示例使用MySQL驱动
	"gorm.io/gorm"

	"oj/internal/problem/domain"
)

var (
	ErrProblemNotFound = errors.New("problem not found")
	ErrTagExists       = errors.New("tag already exists")
	ErrNoTags          = errors.New("no tags found")
)

type ProblemDao interface {
	CreateProblem(ctx context.Context, problem domain.Problem) error
	UpdateProblem(ctx context.Context, id uint64, problem domain.Problem) (domain.Problem, error)
	FindAllProblems(ctx context.Context) ([]domain.Problem, error)
	CreateTag(ctx context.Context, tag string) error
	UpdateTag(ctx context.Context, id uint64, newTag string) error

	FindAllTags(ctx context.Context) ([]domain.Tag, error)
	FindCountInTag(ctx context.Context) ([]domain.TagWithCount, error)
	FindProblemsByName(ctx context.Context, name string) ([]domain.RoughProblem, error)
	FindByTitle(ctx context.Context, tag, title string) (domain.Problem, error)

	getDifficulty(diff uint8) string
	setDifficulty(diff string) uint8
	GetPrompt(p string) []string
	SetPrompt(prompt []string) string
}

type GormProblemDao struct {
	db *gorm.DB
}

func NewProblemDao(db *gorm.DB) ProblemDao {
	return &GormProblemDao{
		db: db,
	}
}

func (dao *GormProblemDao) SetPrompt(prompt []string) string {
	data, err := json.Marshal(prompt)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

func (dao *GormProblemDao) GetPrompt(p string) []string {
	var prompt []string
	err := json.Unmarshal([]byte(p), &prompt)
	if err != nil {
		log.Fatal(err)
	}
	return prompt
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

func (dao *GormProblemDao) CreateProblem(ctx context.Context, problem domain.Problem) error {
	now := time.Now().UnixMilli()

	pm := Problem{
		Title:      problem.Title,
		Content:    problem.Content,
		Difficulty: dao.setDifficulty(string(problem.Difficulty)),
		UserId:     problem.UserId,
		PassRate:   problem.PassRate,
		Prompt:     dao.SetPrompt(problem.Prompt),
		Ctime:      now,
		Uptime:     now,
	}

	err := dao.db.WithContext(ctx).Create(&pm).Error
	return err
}

func (dao *GormProblemDao) UpdateProblem(ctx context.Context, id uint64, problem domain.Problem) (domain.Problem, error) {
	// 使用 GORM 的 Model 进行部分更新
	updateData := make(map[string]interface{})

	// 检查需要更新的字段并添加到 updateData
	if problem.Title != "" {
		updateData["title"] = problem.Title
	}
	if problem.Content != "" {
		updateData["content"] = problem.Content
	}
	if string(problem.Difficulty) != "" {
		updateData["difficulty"] = problem.Difficulty
	}

	if len(updateData) == 0 {
		return domain.Problem{}, errors.New("no fields to update")
	}

	var pm Problem
	result := dao.db.WithContext(ctx).Where("id = ?", id).First(&pm)
	if result.Error != nil {
		return domain.Problem{}, result.Error
	}

	// 更新数据
	if err := dao.db.WithContext(ctx).Model(&pm).Updates(updateData).Error; err != nil {
		return domain.Problem{}, err
	}

	updatePm := domain.Problem{
		Id:         pm.ID,
		UserId:     pm.UserId,
		Title:      pm.Title,
		Content:    pm.Content,
		Prompt:     dao.GetPrompt(pm.Prompt),
		PassRate:   pm.PassRate,
		MaxRuntime: pm.MaxRuntime,
		MaxMem:     pm.MaxMem,
		Difficulty: domain.Difficulty(dao.getDifficulty(pm.Difficulty)),
	}

	return updatePm, nil
}

func (dao *GormProblemDao) FindAllProblems(ctx context.Context) ([]domain.Problem, error) {
	var pms []domain.Problem

	err := dao.db.WithContext(ctx).Model(&domain.Problem{}).Find(&pms).Error

	if err != nil {
		return []domain.Problem{}, err
	}

	return pms, nil
}

func (dao *GormProblemDao) CreateTag(ctx context.Context, tag string) error {
	tg := Tag{
		Name: tag,
	}

	if err := dao.db.WithContext(ctx).Create(&tg).Error; err != nil {
		// 检查是否是唯一约束冲突的错误
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrTagExists
		}
		return err
	}
	return nil
}

func (dao *GormProblemDao) UpdateTag(ctx context.Context, id uint64, newTag string) error {
	var tag Tag

	if err := dao.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error; err != nil {
		return err
	}

	if err := dao.db.WithContext(ctx).Model(&tag).Update("name", newTag).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrTagExists
		}
		return err
	}

	return nil
}

func (dao *GormProblemDao) FindAllTags(ctx context.Context) ([]domain.Tag, error) {
	var tags []domain.Tag

	result := dao.db.WithContext(ctx).Model(&Tag{}).Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}

	// 如果没有任何标签，返回业务逻辑错误
	if len(tags) == 0 {
		return nil, ErrNoTags
	}

	return tags, nil
}

func (dao *GormProblemDao) FindCountInTag(ctx context.Context) ([]domain.TagWithCount, error) {
	var tags []domain.TagWithCount

	result := dao.db.WithContext(ctx).Raw(`
    	SELECT t.id AS tag_id, t.name AS tag_name, COUNT(pt.problem_id) AS problem_count
    	FROM tag t
    	LEFT JOIN problem_tag pt ON t.id = pt.tag_id
    	GROUP BY t.id, t.name
	`).Scan(&tags)

	if result.Error != nil {
		return nil, result.Error
	}

	// 如果没有任何标签，返回业务逻辑错误
	if len(tags) == 0 {
		return nil, ErrNoTags
	}

	return tags, nil
}

func (dao *GormProblemDao) FindProblemsByName(ctx context.Context, name string) ([]domain.RoughProblem, error) {
	var problems []domain.RoughProblem

	query := `
        SELECT p.id, p.tag, p.title, p.pass_rate
        FROM problem p
        JOIN problem_tag pt ON p.id = pt.problem_id
        JOIN tag t ON pt.tag_id = t.id
        WHERE t.name = ?
    `

	err := dao.db.WithContext(ctx).Raw(query, name).Scan(&problems).Error
	if err != nil {
		return nil, err
	}

	return problems, nil
}

func (dao *GormProblemDao) FindByTitle(ctx context.Context, tag, title string) (domain.Problem, error) {
	var problem Problem

	err := dao.db.WithContext(ctx).Where("title = ?", title).First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Problem{}, ErrProblemNotFound
		}

		return domain.Problem{}, err
	}

	pm := domain.Problem{
		Id:         problem.ID,
		UserId:     problem.UserId,
		Title:      problem.Title,
		Content:    problem.Content,
		Tag:        tag,
		Prompt:     dao.GetPrompt(problem.Prompt),
		PassRate:   problem.PassRate,
		MaxMem:     problem.MaxMem,
		MaxRuntime: problem.MaxRuntime,
		Difficulty: domain.Difficulty(dao.getDifficulty(problem.Difficulty)),
	}

	return pm, nil
}