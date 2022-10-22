package tgapi

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func hash(u Update) {
	if u.Message != nil {
		b, _ := json.Marshal(u.Message)
		h := md5.Sum(b)
		u.Message.UUID = hex.EncodeToString(h[:])
	}
	if u.CallbackQuery != nil {
		b, _ := json.Marshal(u.CallbackQuery)
		h := md5.Sum(b)
		u.CallbackQuery.UUID = hex.EncodeToString(h[:])
	}
}
