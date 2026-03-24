package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Saku0512/specter/config"
	"github.com/gin-gonic/gin"
)

func hasStoreOp(rt config.Route) bool {
	return rt.StorePush != "" || rt.StoreList != "" || rt.StoreGet != "" ||
		rt.StorePut != "" || rt.StorePatch != "" || rt.StoreDelete != "" || rt.StoreClear != ""
}

func handleStoreOp(c *gin.Context, rt config.Route, bodyBytes []byte, store *DataStore) {
	key := rt.StoreKey
	if key == "" {
		key = "id"
	}
	id := c.Param(key)

	switch {
	case rt.StorePush != "":
		var item map[string]any
		json.Unmarshal(bodyBytes, &item) //nolint:errcheck
		stored := store.Push(rt.StorePush, item)
		c.JSON(http.StatusCreated, stored)

	case rt.StoreList != "":
		items := store.List(rt.StoreList)

		// Filter: query params not prefixed with _ are treated as field equality filters
		for k, vs := range c.Request.URL.Query() {
			if strings.HasPrefix(k, "_") {
				continue
			}
			want := vs[0]
			filtered := items[:0:0]
			for _, item := range items {
				if fmt.Sprint(item[k]) == want {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}

		// Sort
		if sortField := c.Query("_sort"); sortField != "" {
			desc := strings.ToLower(c.DefaultQuery("_order", "asc")) == "desc"
			sort.SliceStable(items, func(i, j int) bool {
				a := fmt.Sprint(items[i][sortField])
				b := fmt.Sprint(items[j][sortField])
				if desc {
					return a > b
				}
				return a < b
			})
		}

		// Pagination
		if off, err := strconv.Atoi(c.Query("_offset")); err == nil && off > 0 {
			if off >= len(items) {
				items = []map[string]any{}
			} else {
				items = items[off:]
			}
		}
		if lim, err := strconv.Atoi(c.Query("_limit")); err == nil && lim >= 0 && lim < len(items) {
			items = items[:lim]
		}

		c.JSON(http.StatusOK, items)

	case rt.StoreGet != "":
		item, ok := store.Get(rt.StoreGet, id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, item)

	case rt.StorePut != "":
		var item map[string]any
		json.Unmarshal(bodyBytes, &item) //nolint:errcheck
		if item == nil {
			item = map[string]any{}
		}
		store.Put(rt.StorePut, id, item)
		c.JSON(http.StatusOK, item)

	case rt.StorePatch != "":
		var partial map[string]any
		json.Unmarshal(bodyBytes, &partial) //nolint:errcheck
		updated, ok := store.Patch(rt.StorePatch, id, partial)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, updated)

	case rt.StoreDelete != "":
		if !store.Delete(rt.StoreDelete, id) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Status(http.StatusNoContent)

	case rt.StoreClear != "":
		store.Clear(rt.StoreClear)
		c.Status(http.StatusNoContent)
	}
}
