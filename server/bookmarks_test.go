package main

import (
	"encoding/json"
	"testing"

	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//nolint
func makePlugin(api *plugintest.API) *Plugin {
	p := &Plugin{}
	p.SetAPI(api)
	return p
}

func TestStoreBookmarks(t *testing.T) {
	api := makeAPIMock()
	api.On("KVSet", mock.Anything, mock.Anything).Return(nil)
	p := makePlugin(api)

	// initialize test Bookmarks
	u1 := "userID1"
	// u2 := "userID2"

	b1 := &Bookmark{PostID: "ID1", Title: "Title1"}
	b2 := &Bookmark{PostID: "ID2", Title: "Title2"}

	// Add Bookmarks
	bmarks := NewBookmarksWithUser(p.API, u1)
	err := bmarks.add(b1)
	assert.Nil(t, err)
	err = bmarks.add(b2)
	assert.Nil(t, err)

	// Markshal the bmarks and mock api call
	jsonBookmarks, err := json.Marshal(bmarks)
	assert.Nil(t, err)
	api.On("KVSet", "bookmarks_userID1", jsonBookmarks).Return(nil)

	// store bmarks using API
	err = bmarks.storeBookmarks()
	assert.Nil(t, err)
}

func TestAddBookmark(t *testing.T) {
	api := makeAPIMock()
	p := makePlugin(api)

	// create some test bookmarks
	b1 := &Bookmark{PostID: "ID1", Title: "Title1"}
	b2 := &Bookmark{
		PostID:   "ID2",
		Title:    "Title2",
		LabelIDs: []string{"UUID1", "UUID2"},
	}
	b3 := &Bookmark{PostID: "ID3", Title: "Title3"}

	// User 1 has no bookmarks
	u1 := "userID1"
	bmarksU1 := NewBookmarksWithUser(p.API, u1)

	api.On("KVSet", mock.Anything, mock.Anything).Return(nil)
	// User 2 has 2 existing bookmarks
	u2 := "userID2"
	bmarksU2 := NewBookmarksWithUser(p.API, u2)
	err := bmarksU2.add(b1)
	assert.Nil(t, err)
	err = bmarksU2.add(b2)
	assert.Nil(t, err)

	tests := []struct {
		name    string
		userID  string
		bmarks  *Bookmarks
		want    int
		wantErr bool
	}{
		{
			name:    "u1 no previous bookmarks  add one bookmark",
			userID:  u1,
			bmarks:  bmarksU1,
			wantErr: true,
			want:    1,
		},
		{
			name:    "u2 two previous bookmarks  add one bookmark",
			userID:  u2,
			bmarks:  bmarksU2,
			wantErr: true,
			want:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBookmarks, err := json.Marshal(tt.bmarks)
			assert.Nil(t, err)

			key := getBookmarksKey(tt.userID)
			api.On("KVSet", key, mock.Anything).Return(nil)
			api.On("KVGet", key).Return(jsonBookmarks, nil)

			// store bmarks using API
			// bmarks, err := p.addBookmark(tt.userID, b3)
			err = tt.bmarks.addBookmark(b3)
			assert.Nil(t, err)
			assert.Equal(t, tt.want, len(tt.bmarks.ByID))
		})
	}
}

func TestDeleteBookmark(t *testing.T) {
	api := makeAPIMock()
	api.On("KVSet", mock.Anything, mock.Anything).Return(nil)
	p := makePlugin(api)

	// create some test bookmarks
	b1 := &Bookmark{PostID: "ID1", Title: "Title1"}
	b2 := &Bookmark{
		PostID:   "ID2",
		Title:    "Title2",
		LabelIDs: []string{"UUID1", "UUID2"},
	}

	// User 1 has no bookmarks
	u1 := "userID1"
	bmarksU1 := NewBookmarksWithUser(p.API, u1)

	// User 2 has 2 existing bookmarks
	u2 := "userID2"
	bmarksU2 := NewBookmarksWithUser(p.API, u2)
	err := bmarksU2.add(b1)
	assert.Nil(t, err)
	err = bmarksU2.add(b2)
	assert.Nil(t, err)

	tests := []struct {
		name       string
		userID     string
		bmarks     *Bookmarks
		wantErrMsg string
		wantErr    bool
	}{
		{
			name:       "u1 no previous bookmarks  Error out",
			userID:     u1,
			bmarks:     bmarksU1,
			wantErr:    true,
			wantErrMsg: "Bookmark `ID2` does not exist",
		},
		{
			name:       "u2 two previous bookmarks  delete one bookmark",
			userID:     u2,
			bmarks:     bmarksU2,
			wantErr:    false,
			wantErrMsg: "Bookmark `ID2` does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBookmarks, err := json.Marshal(tt.bmarks)
			assert.Nil(t, err)

			key := getBookmarksKey(tt.userID)
			api.On("KVSet", key, mock.Anything).Return(nil)
			api.On("KVGet", key).Return(jsonBookmarks, nil)

			// store bmarks using API
			_, err = tt.bmarks.deleteBookmark(b2.PostID)
			if tt.wantErr {
				assert.Equal(t, err.Error(), tt.wantErrMsg)
				return
			}
			assert.Nil(t, err)
		})
	}
}
