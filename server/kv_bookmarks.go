package main

import (
	"encoding/json"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

// Bookmarks contains a map of bookmarks
type Bookmarks struct {
	ByID   map[string]*Bookmark
	api    plugin.API
	userID string
}

// Bookmark contains information about an individual bookmark
type Bookmark struct {
	PostID     string   `json:"postid"`              // PostID is the ID for the bookmarked post and doubles as the Bookmark ID
	Title      string   `json:"title,omitempty"`     // Title given to the bookmark
	CreateAt   int64    `json:"create_at"`           // The original creation time of the bookmark
	ModifiedAt int64    `json:"update_at"`           // The original creation time of the bookmark
	LabelIDs   []string `json:"label_ids,omitempty"` // Array of labels added to the bookmark
}

// NewBookmarksWithUser returns an initialized Labels for a User
func NewBookmarksWithUser(api plugin.API, userID string) *Bookmarks {
	return &Bookmarks{
		ByID:   make(map[string]*Bookmark),
		api:    api,
		userID: userID,
	}
}

func (b *Bookmarks) add(bmark *Bookmark) error {
	b.ByID[bmark.PostID] = bmark
	if err := b.storeBookmarks(); err != nil {
		return errors.Wrap(err, "failed to add bookmark")
	}
	return nil
}

func (b *Bookmarks) get(bmarkID string) *Bookmark {
	return b.ByID[bmarkID]
}

func (b *Bookmarks) delete(bmarkID string) {
	delete(b.ByID, bmarkID)
}

func (b *Bookmarks) exists(bmarkID string) (*Bookmark, bool) {
	if bmark, ok := b.ByID[bmarkID]; ok {
		return bmark, true
	}
	return nil, false
}

func (b *Bookmarks) updateTimes(bmarkID string) *Bookmark {
	bmark := b.get(bmarkID)
	if bmark.CreateAt == 0 {
		bmark.CreateAt = model.GetMillis()
		bmark.ModifiedAt = bmark.CreateAt
	}
	bmark.ModifiedAt = model.GetMillis()
	return bmark
}

func (b *Bookmarks) updateLabels(bmark *Bookmark) *Bookmark {
	bmarkOrig := b.get(bmark.PostID)
	bmarkOrig.addLabelIDs(bmark.getLabelIDs())
	return bmark
}

func (b *Bookmark) hasUserTitle(bmark *Bookmark) bool {
	return bmark.getTitle() != ""
}

func (b *Bookmark) hasLabels(bmark *Bookmark) bool {
	return bmark.getLabelIDs() != nil
}

func (b *Bookmark) getTitle() string {
	return b.Title
}

func (b *Bookmark) setTitle(title string) {
	b.Title = title
}

func (b *Bookmark) getLabelIDs() []string {
	return b.LabelIDs
}

func (b *Bookmark) addLabelIDs(ids []string) {
	b.LabelIDs = ids
}

// storeBookmarks stores all the users bookmarks
func (b *Bookmarks) storeBookmarks() error {
	jsonBookmarks, jsonErr := json.Marshal(b)
	if jsonErr != nil {
		return jsonErr
	}

	key := getBookmarksKey(b.userID)
	appErr := b.api.KVSet(key, jsonBookmarks)
	if appErr != nil {
		return appErr
	}

	return nil
}
