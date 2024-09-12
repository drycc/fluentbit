package main

import (
	"C"
	"unsafe"

	"context"
	"sort"
	"strings"

	"github.com/fluent/fluent-bit-go/output"
	redis "github.com/redis/go-redis/v9"
)
import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	Stream            string
	MaxLen            int64
	Revision          string
	BuildDate         string
	ControllerName    string
	ControllerRegex   *regexp.Regexp
	ExcludeNamespaces []string
	ClusterClient     *redis.ClusterClient
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	fmt.Printf("Drycc output version %s %s", Revision, BuildDate)
	return output.FLBPluginRegister(ctx, "drycc", "Ship fluent-bit logs to redis xstream")
}

// (fluentbit will call this)
// ctx (context) pointer to fluentbit context (state/ c code)
//
//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {
	var err error
	addrs := strings.Split(output.FLBPluginConfigKey(ctx, "Addrs"), ",")
	sort.Strings(addrs)
	Stream = output.FLBPluginConfigKey(ctx, "Stream")
	MaxLen, err = strconv.ParseInt(output.FLBPluginConfigKey(ctx, "Max_Len"), 10, 64)
	if err != nil {
		MaxLen = 1000
	}
	username := output.FLBPluginConfigKey(ctx, "Username")
	password := output.FLBPluginConfigKey(ctx, "Password")
	ControllerName = output.FLBPluginConfigKey(ctx, "Controller_Name")
	ControllerRegex = regexp.MustCompile(output.FLBPluginConfigKey(ctx, "Controller_Regex"))
	ExcludeNamespaces = strings.Split(output.FLBPluginConfigKey(ctx, "Exclude_Namespaces"), ",")
	ClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		ClusterSlots: func(context.Context) ([]redis.ClusterSlot, error) {
			const slotsSize = 16383
			var size = len(addrs)
			var slotsRange = slotsSize / size
			var slots []redis.ClusterSlot
			for index, addr := range addrs {
				start := slotsRange * index
				end := start + slotsRange
				if (slotsSize - end) < slotsRange {
					end = slotsSize
				}
				slots = append(slots, redis.ClusterSlot{
					Start: start,
					End:   end,
					Nodes: []redis.ClusterNode{{Addr: addr}},
				})
			}
			return slots, nil
		},
		Username:      username,
		Password:      password, // "" == no password
		RouteRandomly: true,
	})
	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, _ *C.char) int {
	status := output.FLB_OK
	context, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	decoder := output.NewDecoder(data, int(length))
	pipeline := ClusterClient.Pipeline()
	for {
		ret, ts, rec := output.GetRecord(decoder)
		if ret != 0 {
			break
		}
		// Get timestamp
		rec["time"] = toTimestamp(ts)
		if !checkRecord(rec) {
			continue
		}
		if values, err := toValues(rec); err == nil {
			if err := pipeline.XAdd(context, &redis.XAddArgs{
				Stream:     Stream,
				NoMkStream: false,
				MaxLen:     MaxLen,
				Approx:     true,
				ID:         "*",
				Values:     values,
			}).Err(); err != nil {
				status = output.FLB_ERROR
			}
		} else {
			status = output.FLB_ERROR
		}
	}
	if _, err := pipeline.Exec(context); err != nil {
		status = output.FLB_ERROR
	}
	return status
}

//export FLBPluginExit
func FLBPluginExit() int {
	ClusterClient.Close()
	return output.FLB_OK
}

func toMap(values map[interface{}]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range values {
		switch t := v.(type) {
		case []byte:
			// prevent encoding to base64
			m[k.(string)] = string(t)
		case map[interface{}]interface{}:
			m[k.(string)] = toMap(t)
		default:
			m[k.(string)] = v
		}
	}
	return m
}

func toValues(values map[interface{}]interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(toMap(values))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data, "timestamp": time.Now().Unix()}, err
}

func toTimestamp(ts interface{}) string {
	var timestamp time.Time
	switch t := ts.(type) {
	case output.FLBTime:
		timestamp = ts.(output.FLBTime).Time
	case uint64:
		timestamp = time.Unix(int64(t), 0)
	default:
		timestamp = time.Now()
	}
	return timestamp.Format(time.RFC3339Nano)
}

func checkRecord(record map[interface{}]interface{}) bool {
	if kubernetes, ok := record["kubernetes"].(map[interface{}]interface{}); ok {
		// drycc controller
		container := fmt.Sprintf("%s", kubernetes["container_name"])
		if ok && strings.HasPrefix(container, ControllerName) {
			log := fmt.Sprintf("%s", record["log"])
			if ok && len(ControllerRegex.FindStringSubmatch(log)) > 0 {
				return true
			}
		}
	}
	return false
}

func main() {}
