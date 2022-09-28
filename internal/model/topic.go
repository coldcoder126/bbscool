package model

import "bbs/pkg/global"

type Topic struct {
	Title     string
	Desc      string
	CreatedBy int64
	Scope     int64
	BaseModel
}

func (Topic) TableName() string {
	return "topic"
}

func (t *Topic) Insert() error {
	return global.Db.Create(t).Error()
}

func (t *Topic) GetById() (topic *Topic, err error) {
	err = global.Db.First(t).Error()
	return t, err
}

func (t *Topic) Delete() error {
	return global.Db.Delete(t).Error()
}

func (t *Topic) UpdateDesc() error {
	return global.Db.Model(t).Update("desc", t.Desc).Error()
}
