package storage

import brokerstorage "github.com/nilebox/broker-server/pkg/stateful/storage"

var _ brokerstorage.StorageWithLease = &postgresStorage{}
