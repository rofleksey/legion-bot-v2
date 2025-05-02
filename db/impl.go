package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"go.etcd.io/bbolt"
)

type Impl struct {
	db *bbolt.DB
}

func NewDatabase(filePath string) (DB, error) {
	db, err := bbolt.Open(filePath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("channels"))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create buckets: %w", err)
	}

	return &Impl{db: db}, nil
}

func (db *Impl) Close() {
	db.db.Close()
}

func (db *Impl) UpdateState(channel string, callback func(state *ChannelState)) {
	err := db.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("channels"))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		var state ChannelState
		data := bucket.Get([]byte(channel))
		if data != nil {
			if err := json.Unmarshal(data, &state); err != nil {
				return err
			}
		} else {
			state = NewChannelState(channel)
		}

		callback(&state)

		data, err := json.Marshal(state)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(channel), data)
	})

	if err != nil {
		slog.Error("Failed to update state for channel",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
	}
}

func (db *Impl) GetState(channel string) ChannelState {
	var state ChannelState

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("channels"))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		data := bucket.Get([]byte(channel))
		if data != nil {
			if err := json.Unmarshal(data, &state); err != nil {
				return err
			}
		} else {
			state = NewChannelState(channel)
		}

		return nil
	})

	if err != nil {
		slog.Error("Failed to get state for channel",
			slog.String("channel", channel),
			slog.Any("error", err),
		)
		return ChannelState{}
	}

	return state
}

func (db *Impl) GetAllStates() []ChannelState {
	var states []ChannelState

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("channels"))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var state ChannelState
			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}
			states = append(states, state)
			return nil
		})
	})

	if err != nil {
		slog.Error("Failed to get all channel states",
			slog.Any("error", err),
		)
		return nil
	}

	return states
}

func (db *Impl) ReadAllStates(callback func(state *ChannelState)) {
	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("channels"))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		return bucket.ForEach(func(k, v []byte) error {
			var state ChannelState
			if err := json.Unmarshal(v, &state); err != nil {
				return err
			}
			callback(&state)
			return nil
		})
	})

	if err != nil {
		slog.Error("Failed to iterate channel states",
			slog.Any("error", err),
		)
	}
}

func (db *Impl) GetAllChannelNames() []string {
	var channelNames []string

	err := db.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("channels"))
		if bucket == nil {
			return errors.New("bucket not found")
		}

		return bucket.ForEach(func(k, _ []byte) error {
			channelNames = append(channelNames, string(k))
			return nil
		})
	})

	if err != nil {
		slog.Error("Failed to get channel names",
			slog.Any("error", err),
		)
		return nil
	}

	return channelNames
}
