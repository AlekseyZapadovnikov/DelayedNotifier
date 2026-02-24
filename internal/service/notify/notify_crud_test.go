package notify_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/notify"
	projectmocks "github.com/AlekseyZapadovnikov/DelayedNotifier/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifier_CreateNotify(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		rec := &models.Record{Id: 101}

		cache.EXPECT().Add(ctx, rec).Return(nil)

		err := n.CreateNotify(ctx, rec)

		require.NoError(t, err)
	})

	t.Run("cache error", func(t *testing.T) {
		t.Parallel()

		cacheErr := errors.New("cache add failed")
		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		rec := &models.Record{Id: 102}

		cache.EXPECT().Add(ctx, rec).Return(cacheErr)

		err := n.CreateNotify(ctx, rec)

		require.Error(t, err)
		assert.ErrorIs(t, err, cacheErr)
	})
}

func TestNotifier_GetNotifyStatByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		id := int64(201)
		rec := &models.Record{Id: id, RecStat: models.RecordStatusSended}

		cache.EXPECT().GetByID(ctx, id).Return(rec, nil)

		status, err := n.GetNotifyStatByID(ctx, id)

		require.NoError(t, err)
		assert.Equal(t, models.RecordStatusSended, status)
	})

	t.Run("cache error", func(t *testing.T) {
		t.Parallel()

		cacheErr := errors.New("cache get failed")
		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		id := int64(202)

		cache.EXPECT().GetByID(ctx, id).Return(nil, cacheErr)

		status, err := n.GetNotifyStatByID(ctx, id)

		require.Error(t, err)
		assert.Empty(t, status)
		assert.ErrorIs(t, err, cacheErr)
	})

	t.Run("nil record without error", func(t *testing.T) {
		t.Parallel()

		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		id := int64(203)

		cache.EXPECT().GetByID(ctx, id).Return(nil, nil)

		status, err := n.GetNotifyStatByID(ctx, id)

		require.Error(t, err)
		assert.Empty(t, status)
		assert.ErrorIs(t, err, notify.ErrNotificationRecordNotFound)
	})
}

func TestNotifier_DeleteNotifyByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		id := int64(301)

		cache.EXPECT().DeleteByID(ctx, id).Return(nil)

		err := n.DeleteNotifyByID(ctx, id)

		require.NoError(t, err)
	})

	t.Run("cache error", func(t *testing.T) {
		t.Parallel()

		cacheErr := errors.New("cache delete failed")
		cache := projectmocks.NewMockNotificationCache(t)
		n := notify.NewNotifierWithSender(nil, cache, nil, config.NotificationConfig{})

		ctx := context.Background()
		id := int64(302)

		cache.EXPECT().DeleteByID(ctx, id).Return(cacheErr)

		err := n.DeleteNotifyByID(ctx, id)

		require.Error(t, err)
		assert.ErrorIs(t, err, cacheErr)
	})
}
