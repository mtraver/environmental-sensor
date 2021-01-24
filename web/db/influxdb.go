package db

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/mtraver/environmental-sensor/measurement"
	mpb "github.com/mtraver/environmental-sensor/measurementpb"
)

func newInfluxDBPoints(m *mpb.Measurement) ([]*write.Point, error) {
	sm, err := measurement.NewStorableMeasurement(m)
	if err != nil {
		return nil, err
	}

	vm := sm.ValueMap()
	points := make([]*write.Point, 0, len(vm))
	for name, v := range vm {
		p := influxdb2.NewPointWithMeasurement("stat")
		if metric, ok := measurement.GetMetric(name); ok {
			p = p.AddField(metric.Abbrv, v)
		} else {
			p = p.AddField(name, v)
		}

		points = append(points, p.AddTag("device", sm.DeviceID).SetTime(sm.Timestamp))
	}

	return points, nil
}

type InfluxDB struct {
	serverURL string
	token     string
	org       string
	bucket    string
}

func NewInfluxDB(serverURL, token, org, bucket string) *InfluxDB {
	return &InfluxDB{
		serverURL: serverURL,
		token:     token,
		org:       org,
		bucket:    bucket,
	}
}

func (db *InfluxDB) Save(ctx context.Context, m *mpb.Measurement) error {
	points, err := newInfluxDBPoints(m)
	if err != nil {
		return err
	}

	client := influxdb2.NewClient(db.serverURL, db.token)
	defer client.Close()

	writeAPI := client.WriteAPI(db.org, db.bucket)
	for _, p := range points {
		writeAPI.WritePoint(p)
	}
	writeAPI.Flush()

	return nil
}
