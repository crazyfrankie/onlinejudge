package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/crazyfrankie/onlinejudge/internal/judgement/domain"
)

type SubmissionDao interface {
	CreateSubmit(ctx context.Context, sub domain.Submission) (uint64, error)
	CreateEvaluate(ctx context.Context, eva domain.Evaluation) error
	UpdateEvaluate(ctx context.Context, pid, sid uint64, state string) error
	FindEvaluate(ctx context.Context, sid uint64) (domain.Evaluation, error)
}

type SubmitDao struct {
	db *gorm.DB
}

func NewSubmitDao(db *gorm.DB) *SubmitDao {
	return &SubmitDao{db: db}
}

func (d *SubmitDao) CreateSubmit(ctx context.Context, sub domain.Submission) (uint64, error) {
	now := time.Now().Unix()
	submit := &Submission{
		ProblemID:  sub.ProblemID,
		UserId:     sub.UserId,
		Code:       sub.Code,
		Language:   sub.Language,
		SubmitTime: sub.SubmitTime,
		Ctime:      now,
		Uptime:     now,
	}
	err := d.db.WithContext(ctx).Create(submit).Error
	if err != nil {
		return 0, err
	}

	return submit.Id, nil
}

func (d *SubmitDao) CreateEvaluate(ctx context.Context, eva domain.Evaluation) error {
	now := time.Now().Unix()
	var st State
	err := d.db.WithContext(ctx).Create(&Evaluation{
		SubmissionId: eva.SubmissionId,
		ProblemId:    eva.ProblemId,
		Lang:         eva.Lang,
		CpuTimeUsed:  eva.CpuTimeUsed,
		RealTimeUsed: eva.RealTimeUsed,
		MemoryUsed:   eva.MemoryUsed,
		StatusMsg:    eva.StatusMsg,
		State:        st.toUint8(eva.State),
		Ctime:        now,
		Utime:        now,
	}).Error
	if err != nil {
		return err
	}

	return nil
}

func (d *SubmitDao) UpdateEvaluate(ctx context.Context, pid, sid uint64, state string) error {
	var st State
	err := d.db.WithContext(ctx).Model(&Evaluation{}).
		Where("problem_id = ? AND submission_id = ?", pid, sid).
		Update("state", st.toUint8(state)).Error
	if err != nil {
		return err
	}

	return nil
}

func (d *SubmitDao) FindEvaluate(ctx context.Context, sid uint64) (domain.Evaluation, error) {
	var eva Evaluation
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sub Submission
		err := d.db.WithContext(ctx).Model(&Submission{}).
			Where("id = ?", sid).Find(&sub).Error
		if err != nil {
			return err
		}

		err = d.db.WithContext(ctx).Model(&Evaluation{}).
			Where("problem_id = ? AND submission_id = ?", sub.ProblemID, sid).
			Find(&eva).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return domain.Evaluation{}, err
	}

	return domain.Evaluation{
		Id:           eva.Id,
		SubmissionId: eva.SubmissionId,
		ProblemId:    eva.ProblemId,
		Lang:         eva.Lang,
		CpuTimeUsed:  eva.CpuTimeUsed,
		RealTimeUsed: eva.RealTimeUsed,
		MemoryUsed:   eva.MemoryUsed,
		StatusMsg:    eva.StatusMsg,
		State:        State(eva.State).toString(),
	}, err
}
