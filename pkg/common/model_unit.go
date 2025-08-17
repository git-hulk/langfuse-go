package common

import "github.com/hashicorp/go-set/v3"

var ModelUsageUnits = set.From([]string{
	"TOKENS",
	"CHARACTERS",
	"MILLISECONDS",
	"SECONDS",
	"IMAGES",
	"REQUESTS",
})
