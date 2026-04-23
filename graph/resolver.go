package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"github.com/mtraver/environmental-sensor/database"
)

type Resolver struct {
	Database   database.Database
	AWSRegion  string
	AWSRoleARN string
}
