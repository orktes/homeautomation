package bolt

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/orktes/homeautomation/bridge/adapter"
)

type BOLT struct {
	id string
	adapter.Updater

	db *bolt.DB
}

func (b *BOLT) ID() string {
	return b.id
}

func (b *BOLT) Get(id string) (interface{}, error) {
	var res interface{}
	return res, b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(b.id))
		if bucket == nil {
			return nil
		}

		d := bucket.Get([]byte(id))

		if len(d) > 0 {
			if err := json.Unmarshal(d, &res); err != nil {
				return err
			}
		}

		return nil
	})
}

func (b *BOLT) Set(id string, val interface{}) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		d, err := json.Marshal(val)
		if err != nil {
			return err
		}
		bucket, err := tx.CreateBucketIfNotExists([]byte(b.id))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(id), d)
	})

	if err == nil {
		b.Updater.SendUpdate(adapter.Update{
			ValueContainer: b,
			Updates: []adapter.ValueUpdate{
				adapter.ValueUpdate{
					Key:   b.ID() + "/" + id,
					Value: val,
				},
			},
		})
	}

	return err
}

func (b *BOLT) GetAll() (map[string]interface{}, error) {
	vals := map[string]interface{}{}
	// TODO figure out if this is wise or reasonable
	return vals, nil
}

func (b *BOLT) UpdateChannel() <-chan adapter.Update {
	return b.Updater.UpdateChannel()
}

func (b *BOLT) Close() error {
	return b.db.Close()
}

// Create returns a new bolt instance
func Create(id string, config map[string]interface{}) (adapter.Adapter, error) {

	// TODO Allow multiple BOLTS to be backed by same database file but different bucked
	databaseFile := config["database_file"].(string)
	db, err := bolt.Open(databaseFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	b := &BOLT{
		id: id,
		db: db,
	}

	return b, nil

}
