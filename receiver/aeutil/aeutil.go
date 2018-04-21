package aeutil

import (
	"context"
	"strings"

	"google.golang.org/appengine"
)

func GetProjectID(ctx context.Context) string {
	// From the documentation of appengine.AppID:
	//
	//   AppID returns the application ID for the current application. The string
	//   will be a plain application ID (e.g. "appid"), with a domain prefix for
	//   custom domain deployments (e.g. "example.com:appid").
	//
	// Here we just want the app ID (don't care if it's deployed to a custom
	// domain) so split at the first colon. This is fine because an app ID can
	// only have lowercase letters, digits, and hyphens.
	appIDParts := strings.Split(appengine.AppID(ctx), ":")
	return appIDParts[len(appIDParts)-1]
}
