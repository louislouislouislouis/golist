package ui

import (
	"github.com/charmbracelet/bubbles/list"
)

func createDisplayedListFromMetadata[T any](
	slice []T,
	getMetadata func(T) DiplayableItemList,
) []list.Item {
	items := make([]list.Item, len(slice))
	for i, itemm := range slice {
		meta := getMetadata(itemm)
		newItem := displayedItem{
			title: meta.GetName(),
			desc:  meta.GetUID(),
		}
		items[i] = newItem
	}
	return items
}
