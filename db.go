package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

var (
	_client     *redis.Client
	_randSeeded = false
)

// '0' and '1' dropped as they are hard to distinguish from O and I
const asciiCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ23456789"

const (
	LookupSuccess    = iota
	LookupNoSuchCode = iota
	LookupDBError    = iota
)

const (
	InsertSuccess  = iota
	InsertConflict = iota
	InsertDBError  = iota
)

const (
	DeleteSuccess = iota
	DeleteDBError = iota
)

func getRedisClient() *redis.Client {
	if _client != nil {
		return _client
	}

	conf := getConfig()
	_client = redis.NewClient(&redis.Options{
		Addr:     conf.RedisConn,
		Password: "",
		DB:       0,
	})
	return _client
}

func getFullURL(code string) (string, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			return "", LookupNoSuchCode
		default:
			if val, err := client.Get(code).Result(); err == nil {
				return val, LookupSuccess
			}
			return "", LookupDBError
		}
	} else {
		return "", LookupDBError
	}
}

func getCodeMeta(code string) (CodeMetaResponse, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			return CodeMetaResponse{}, LookupNoSuchCode
		default:
			if valMeta, err := client.Get(fmt.Sprintf("meta/%s", code)).Result(); err == nil {
				meta := &CodeMeta{}
				err = json.Unmarshal([]byte(valMeta), meta)
				if err != nil {
					return CodeMetaResponse{}, LookupDBError
				}
				fullURL, err := getFullURL(code)
				if err != LookupSuccess {
					return CodeMetaResponse{}, LookupDBError
				}
				return CodeMetaResponse{
					FullURL: fullURL,
					Meta:    *meta,
				}, LookupSuccess
			}
			return CodeMetaResponse{}, LookupDBError
		}
	} else {
		return CodeMetaResponse{}, LookupDBError
	}
}

func deleteCode(code string) (DeleteResponse, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			return DeleteResponse{
				Code:   code,
				Status: "noexist",
			}, DeleteSuccess
		default:
			if err := client.Del(code).Err(); err == nil {
				client.Del(fmt.Sprintf("meta/%s", code))
				return DeleteResponse{
					Code:   code,
					Status: "deleted",
				}, DeleteSuccess
			}
			return DeleteResponse{}, DeleteDBError
		}
	} else {
		return DeleteResponse{}, DeleteDBError
	}
}

func addURLWithCode(url, code, meta string, user APIKey) (string, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			metaJSON, err := json.Marshal(&CodeMeta{
				Owner:    user.Name,
				Time:     time.Now(),
				UserMeta: meta,
			})
			if err != nil {
				return "", InsertDBError
			}
			if _, err := client.Set(code, url, 0).Result(); err != nil {
				return "", InsertDBError
			}
			client.Set(fmt.Sprintf("meta/%s", code), metaJSON, 0)
			return fmt.Sprintf("%s/%s", getConfig().ServerURL, code), InsertSuccess
		default:
			return "", InsertConflict
		}
	} else {
		return "", InsertDBError
	}
}

func addURL(url, meta string, user APIKey) (string, int) {
	for {
		code := randomASCIISequence(getConfig().CodeLength)
		newURL, errCode := addURLWithCode(url, code, meta, user)
		switch errCode {
		case InsertSuccess:
			return newURL, errCode
		case InsertConflict:
			e.Logger.Warn("Duplicate code in generation. Code length too small? %s", code)
			continue
		case InsertDBError:
			return "", errCode
		default:
			panic(fmt.Sprintf("Unknown return code from addURLWithCode: %v", errCode))
		}
	}
}

func randomASCIISequence(length int) string {
	if !_randSeeded {
		rand.Seed(time.Now().UnixNano())
		_randSeeded = true
	}

	// Based on http://stackoverflow.com/a/31832326
	b := make([]byte, length)
	for i := range b {
		b[i] = asciiCharacters[rand.Intn(len(asciiCharacters))]
	}
	return string(b)
}
