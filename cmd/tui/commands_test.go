package tui

import (
	"errors"
	"testing"

	"github.com/benmatselby/wibble/pkg/dao"
	"github.com/benmatselby/wibble/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestFetchArticles_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	want := []models.Article{{ID: 1, FeedID: 42, Title: "Article One"}}
	db.EXPECT().GetArticlesByFeedID(int64(42)).Return(want, nil)

	cmd := fetchArticles(db, 42)
	msg, ok := cmd().(articlesLoadedMsg)
	if !ok {
		t.Fatalf("expected articlesLoadedMsg, got %T", cmd())
	}
	if msg.err != nil {
		t.Errorf("err = %v, want nil", msg.err)
	}
	if len(msg.articles) != 1 || msg.articles[0].Title != "Article One" {
		t.Errorf("articles = %+v, want %+v", msg.articles, want)
	}
}

func TestFetchArticles_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	wantErr := errors.New("boom")
	db.EXPECT().GetArticlesByFeedID(int64(1)).Return(nil, wantErr)

	cmd := fetchArticles(db, 1)
	msg, ok := cmd().(articlesLoadedMsg)
	if !ok {
		t.Fatalf("expected articlesLoadedMsg, got %T", cmd())
	}
	if !errors.Is(msg.err, wantErr) {
		t.Errorf("err = %v, want %v", msg.err, wantErr)
	}
	if msg.articles != nil {
		t.Errorf("articles = %+v, want nil", msg.articles)
	}
}

func TestFetchFeeds_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	want := []models.Feed{{ID: 1, Title: "Feed One"}}
	db.EXPECT().GetFeeds().Return(want, nil)

	cmd := fetchFeeds(db)
	msg, ok := cmd().(feedsLoadedMsg)
	if !ok {
		t.Fatalf("expected feedsLoadedMsg, got %T", cmd())
	}
	if msg.err != nil {
		t.Errorf("err = %v, want nil", msg.err)
	}
	if len(msg.feeds) != 1 || msg.feeds[0].Title != "Feed One" {
		t.Errorf("feeds = %+v, want %+v", msg.feeds, want)
	}
}

func TestFetchFeeds_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	wantErr := errors.New("boom")
	db.EXPECT().GetFeeds().Return(nil, wantErr)

	cmd := fetchFeeds(db)
	msg, ok := cmd().(feedsLoadedMsg)
	if !ok {
		t.Fatalf("expected feedsLoadedMsg, got %T", cmd())
	}
	if !errors.Is(msg.err, wantErr) {
		t.Errorf("err = %v, want %v", msg.err, wantErr)
	}
	if msg.feeds != nil {
		t.Errorf("feeds = %+v, want nil", msg.feeds)
	}
}

func TestFetchTags_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	want := []models.Tag{{ID: 1, Name: "golang"}}
	db.EXPECT().GetTags().Return(want, nil)

	cmd := fetchTags(db)
	msg, ok := cmd().(tagsLoadedMsg)
	if !ok {
		t.Fatalf("expected tagsLoadedMsg, got %T", cmd())
	}
	if msg.err != nil {
		t.Errorf("err = %v, want nil", msg.err)
	}
	if len(msg.tags) != 1 || msg.tags[0].Name != "golang" {
		t.Errorf("tags = %+v, want %+v", msg.tags, want)
	}
}

func TestFetchTags_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	wantErr := errors.New("boom")
	db.EXPECT().GetTags().Return(nil, wantErr)

	cmd := fetchTags(db)
	msg, ok := cmd().(tagsLoadedMsg)
	if !ok {
		t.Fatalf("expected tagsLoadedMsg, got %T", cmd())
	}
	if !errors.Is(msg.err, wantErr) {
		t.Errorf("err = %v, want %v", msg.err, wantErr)
	}
	if msg.tags != nil {
		t.Errorf("tags = %+v, want nil", msg.tags)
	}
}

func TestFetchArticleTags_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	want := []models.Tag{{ID: 2, Name: "devops"}}
	db.EXPECT().GetTagsForArticle(int64(7)).Return(want, nil)

	cmd := fetchArticleTags(db, 7)
	msg, ok := cmd().(articleTagsLoadedMsg)
	if !ok {
		t.Fatalf("expected articleTagsLoadedMsg, got %T", cmd())
	}
	if msg.err != nil {
		t.Errorf("err = %v, want nil", msg.err)
	}
	if len(msg.tags) != 1 || msg.tags[0].Name != "devops" {
		t.Errorf("tags = %+v, want %+v", msg.tags, want)
	}
}

func TestFetchArticleTags_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	wantErr := errors.New("boom")
	db.EXPECT().GetTagsForArticle(int64(7)).Return(nil, wantErr)

	cmd := fetchArticleTags(db, 7)
	msg, ok := cmd().(articleTagsLoadedMsg)
	if !ok {
		t.Fatalf("expected articleTagsLoadedMsg, got %T", cmd())
	}
	if !errors.Is(msg.err, wantErr) {
		t.Errorf("err = %v, want %v", msg.err, wantErr)
	}
	if msg.tags != nil {
		t.Errorf("tags = %+v, want nil", msg.tags)
	}
}

func TestFetchArticlesByTag_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	want := []models.Article{{ID: 3, Title: "Tagged Article"}}
	db.EXPECT().GetArticlesByTagID(int64(9)).Return(want, nil)

	cmd := fetchArticlesByTag(db, 9)
	msg, ok := cmd().(articlesLoadedMsg)
	if !ok {
		t.Fatalf("expected articlesLoadedMsg, got %T", cmd())
	}
	if msg.err != nil {
		t.Errorf("err = %v, want nil", msg.err)
	}
	if len(msg.articles) != 1 || msg.articles[0].Title != "Tagged Article" {
		t.Errorf("articles = %+v, want %+v", msg.articles, want)
	}
}

func TestFetchArticlesByTag_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := dao.NewMockDaoClient(ctrl)

	wantErr := errors.New("boom")
	db.EXPECT().GetArticlesByTagID(int64(9)).Return(nil, wantErr)

	cmd := fetchArticlesByTag(db, 9)
	msg, ok := cmd().(articlesLoadedMsg)
	if !ok {
		t.Fatalf("expected articlesLoadedMsg, got %T", cmd())
	}
	if !errors.Is(msg.err, wantErr) {
		t.Errorf("err = %v, want %v", msg.err, wantErr)
	}
	if msg.articles != nil {
		t.Errorf("articles = %+v, want nil", msg.articles)
	}
}
