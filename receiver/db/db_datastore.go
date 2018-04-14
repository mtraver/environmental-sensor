package db

import (
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"

	"receiver/measurement"
)

const (
	datastoreKind = "measurement"

	// Datastore queries are limited to this many entities, and multiple queries
	// are made to fetch all results.
	queryLimit = 1000
)

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

func executeQuery(
	ctx context.Context,
	q *datastore.Query) (map[string][]measurement.StorableMeasurement, error) {
	results := make(map[string][]measurement.StorableMeasurement)

	// Don't modify the original query. We'll continue to derive queries from it
	// using a cursor to break apart the whole query into multiple smaller ones.
	derivedQuery := q.Limit(queryLimit)

	for {
		processed := 0

		it := derivedQuery.Run(ctx)
		for {
			var m measurement.StorableMeasurement
			_, err := it.Next(&m)
			if err == datastore.Done {
				cursor, err := it.Cursor()
				if err != nil {
					return make(map[string][]measurement.StorableMeasurement), err
				}

				// The current query finished, so make a new one that starts
				// where it left off.
				derivedQuery = q.Start(cursor).Limit(queryLimit)
				break
			} else if err != nil {
				return make(map[string][]measurement.StorableMeasurement), err
			}

			if _, ok := results[m.DeviceId]; !ok {
				results[m.DeviceId] = []measurement.StorableMeasurement{}
			}
			results[m.DeviceId] = append(results[m.DeviceId], m)

			processed++
		}

		if processed < queryLimit {
			// The last query returned fewer results than the limit, meaning that a
			// subsequent query would return nothing, so we're done.
			break
		}
	}

	return results, nil
}

func (db *datastoreDB) GetMeasurementsSince(
	ctx context.Context,
	startTime time.Time) (map[string][]measurement.StorableMeasurement, error) {
	// Don't need to filter by device ID here because building the map
	// has the effect of sorting by device ID.
	q := datastore.NewQuery(datastoreKind).Filter(
		"timestamp >=", startTime).Order("timestamp")

	return executeQuery(ctx, q)
}

func (db *datastoreDB) GetMeasurementsBetween(
	ctx context.Context, startTime time.Time,
	endTime time.Time) (map[string][]measurement.StorableMeasurement, error) {
	// Don't need to filter by device ID here because building the map
	// has the effect of sorting by device ID.
	q := datastore.NewQuery(datastoreKind).Filter(
		"timestamp >=", startTime).Filter(
		"timestamp <=", endTime).Order("timestamp")

	return executeQuery(ctx, q)
}
