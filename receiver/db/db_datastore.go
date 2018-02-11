package db

import (
  "golang.org/x/net/context"

  "google.golang.org/appengine/datastore"

  "receiver/measurement"
)

const datastoreKind = "measurement"

type datastoreDB struct {
  projectID string
}

// Ensure datastoreDB implements Database
var _ Database = &datastoreDB{}

func NewDatastoreDB(projectID string) Database {
  return &datastoreDB{
      projectID: projectID,
  }
}

func (db *datastoreDB) Save(ctx context.Context,
                            m *measurement.Measurement) error {
  storableMeasurement, err := m.ToStorableMeasurement()
  if err != nil {
    return err
  }

  key := datastore.NewKey(
      ctx, datastoreKind, storableMeasurement.DBKey(), 0, nil)

  // Only store the measurement if it doesn't exist
  err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
    var x measurement.StorableMeasurement
    if err := datastore.Get(ctx, key, &x); err != datastore.ErrNoSuchEntity {
      return err
    }

    _, err := datastore.Put(ctx, key, &storableMeasurement)
    return err
  }, nil)

  return err
}
