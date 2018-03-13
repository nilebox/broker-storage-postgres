package storage

import "github.com/nilebox/broker-server/pkg/stateful/retry"

var _ retry.StorageWithLease = &postgresStorage{}
