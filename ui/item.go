package ui

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DiplayableItemList interface {
	GetName() string
	GetUID() string
}

type displayedItem struct {
	title, desc string
	isSelected  bool
}

func (i displayedItem) Title() string       { return i.title }
func (i displayedItem) Description() string { return i.desc }
func (i displayedItem) FilterValue() string { return i.title }
func (i displayedItem) IsSelected() bool    { return i.isSelected }

type displayableMeta struct {
	meta metaV1.Object
}

type displayableContainer struct {
	v1.Container
}

func (m displayableMeta) GetUID() string {
	return string(m.meta.GetUID())
}

func (m displayableMeta) GetName() string {
	return m.meta.GetName()
}

func (container displayableContainer) GetName() string { return container.Name }

func (container displayableContainer) GetUID() string { return container.Image }
