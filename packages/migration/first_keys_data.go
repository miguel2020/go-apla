package migration

import "github.com/GenesisKernel/go-genesis/packages/consts"

var firstKeysDataSQL = `INSERT INTO "1_keys" (id, pub, blocked, ecosystem) VALUES (` + consts.GuestKey + `, '` + consts.GuestPublic + `', 1, '1');`