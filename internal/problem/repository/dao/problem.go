package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/crazyfrankie/onlinejudge/internal/problem/domain"
)

var (
	ErrProblemNotFound = errors.New("problem not found")
	ErrProblemExists   = errors.New("problem already exists")
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
	FindTestById(ctx context.Context, id uint64) ([]domain.TestCase, error)
	FindAllTestById(ctx context.Context, id uint64) ([]domain.TestCase, string, error)
	FindTestByIdLocal(ctx context.Context, id uint64) (domain.LocalJudge, error)
}

type GormProblemDao struct {
	db *gorm.DB
}

func NewProblemDao(db *gorm.DB) ProblemDao {
	return &GormProblemDao{
		db: db,
	}
}

func (dao *GormProblemDao) CreateProblem(ctx context.Context, problem domain.Problem) error {
	now := time.Now().UnixMilli()

	testCasesJSON, err := sonic.Marshal(problem.TestCases)
	if err != nil {
		return err
	}

	tmpl := fmt.Sprintf(QuestionTemplate, problem.FuncName)
	pm := Problem{
		Title:        problem.Title,
		Content:      problem.Content,
		Difficulty:   problem.Difficulty,
		UserId:       problem.UserId,
		PassRate:     problem.PassRate,
		TestCases:    string(testCasesJSON),
		Params:       problem.Params,
		PreDefine:    problem.PreDefine,
		TemplateCode: tmpl,
		Ctime:        now,
		Uptime:       now,
	}

	err = dao.db.WithContext(ctx).Create(&pm).Error

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrProblemExists
	}

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
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.Problem{}, ErrProblemNotFound
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
		PassRate:   pm.PassRate,
		MaxRuntime: pm.MaxRuntime,
		MaxMem:     pm.MaxMem,
		Difficulty: pm.Difficulty,
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
		PassRate:   problem.PassRate,
		MaxMem:     problem.MaxMem,
		MaxRuntime: problem.MaxRuntime,
		Difficulty: problem.Difficulty,
	}

	return pm, nil
}

func (dao *GormProblemDao) FindTestById(ctx context.Context, id uint64) ([]domain.TestCase, error) {
	var problem Problem

	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.TestCase{}, ErrProblemNotFound
		}
		return []domain.TestCase{}, err
	}

	var testCases []domain.TestCase
	err = sonic.Unmarshal([]byte(problem.TestCases), &testCases)
	if err != nil {
		return testCases[:3], err
	}

	return testCases[:3], nil
}

func (dao *GormProblemDao) FindAllTestById(ctx context.Context, id uint64) ([]domain.TestCase, string, error) {
	var problem Problem

	err := dao.db.WithContext(ctx).Where("id = ?", id).Select("template_code", "params", "test_cases").First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []domain.TestCase{}, "", ErrProblemNotFound
		}
		return []domain.TestCase{}, "", err
	}

	var testCases []domain.TestCase
	err = sonic.Unmarshal([]byte(problem.TestCases), &testCases)
	if err != nil {
		return testCases, "", err
	}

	return testCases, problem.TemplateCode, nil
}

func (dao *GormProblemDao) FindTestByIdLocal(ctx context.Context, id uint64) (domain.LocalJudge, error) {
	var problem Problem

	err := dao.db.WithContext(ctx).Where("id = ?", id).Select("template_code", "params", "test_cases").First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.LocalJudge{}, ErrProblemNotFound
		}
		return domain.LocalJudge{}, err
	}

	var judge domain.LocalJudge
	var tc []domain.LocalTestCase
	err = sonic.Unmarshal([]byte(problem.TestCases), &tc)
	if err != nil {
		fmt.Println(err)
		return domain.LocalJudge{}, err
	}
	judge.Params = problem.Params
	judge.TemplateCode = problem.TemplateCode
	judge.TestCases = tc
	return judge, nil
}
