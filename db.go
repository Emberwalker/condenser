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
	lookupSuccess    = iota
	lookupNoSuchCode = iota
	lookupDBError    = iota
)

const (
	insertSuccess  = iota
	insertConflict = iota
	insertDBError  = iota
)

const (
	deleteSuccess = iota
	deleteDBError = iota
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
			return "", lookupNoSuchCode
		default:
			if val, err := client.Get(code).Result(); err == nil {
				return val, lookupSuccess
			}
			logger.Warnf("DB error: %+v", err)
			return "", lookupDBError
		}
	} else {
		logger.Warnf("DB error: %+v", err)
		return "", lookupDBError
	}
}

func getCodeMeta(code string) (codeMetaResponse, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			return codeMetaResponse{}, lookupNoSuchCode
		default:
			if valMeta, err := client.Get(fmt.Sprintf("meta/%s", code)).Result(); err == nil {
				meta := &codeMeta{}
				err = json.Unmarshal([]byte(valMeta), meta)
				if err != nil {
					logger.Warnf("json unmarshal error: %+v", err)
					return codeMetaResponse{}, lookupDBError
				}
				fullURL, err := getFullURL(code)
				if err != lookupSuccess {
					return codeMetaResponse{}, lookupDBError
				}
				return codeMetaResponse{
					FullURL: fullURL,
					Meta:    *meta,
				}, lookupSuccess
			}
			logger.Warnf("DB error: %+v", err)
			return codeMetaResponse{}, lookupDBError
		}
	} else {
		logger.Warnf("DB error: %+v", err)
		return codeMetaResponse{}, lookupDBError
	}
}

func deleteCode(code string) (deleteResponse, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			return deleteResponse{
				Code:   code,
				Status: "noexist",
			}, deleteSuccess
		default:
			if err := client.Del(code).Err(); err == nil {
				client.Del(fmt.Sprintf("meta/%s", code))
				return deleteResponse{
					Code:   code,
					Status: "deleted",
				}, deleteSuccess
			}
			logger.Warnf("DB error: %+v", err)
			return deleteResponse{}, deleteDBError
		}
	} else {
		logger.Warnf("DB error: %+v", err)
		return deleteResponse{}, deleteDBError
	}
}

func addURLWithCode(url, code, meta string, user APIKey) (string, int) {
	client := getRedisClient()
	code = strings.ToUpper(code)
	if val, err := client.Exists(code).Result(); err == nil {
		switch val {
		case 0:
			metaJSON, err := json.Marshal(&codeMeta{
				Owner:    user.Name,
				Time:     time.Now(),
				UserMeta: meta,
			})
			if err != nil {
				logger.Warnf("json marshal error: %+v", err)
				return "", insertDBError
			}
			if _, err := client.Set(code, url, 0).Result(); err != nil {
				logger.Warnf("DB error: %+v", err)
				return "", insertDBError
			}
			client.Set(fmt.Sprintf("meta/%s", code), metaJSON, 0)
			return fmt.Sprintf("%s/%s", getConfig().ServerURL, code), insertSuccess
		default:
			return "", insertConflict
		}
	} else {
		logger.Warnf("DB error: %+v", err)
		return "", insertDBError
	}
}

func addURL(url, meta string, user APIKey) (string, int) {
	for {
		code := randomASCIISequence(getConfig().CodeLength)
		newURL, errCode := addURLWithCode(url, code, meta, user)
		switch errCode {
		case insertSuccess:
			return newURL, errCode
		case insertConflict:
			logger.Warnf("Duplicate code in generation. Code length too small? %s", code)
			continue
		case insertDBError:
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
