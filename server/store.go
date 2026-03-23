package server

import (
	"fmt"
	"sync"
)

// DataStore is a named-collection in-memory CRUD store.
// Each collection holds objects identified by a string "id" field.
type DataStore struct {
	mu          sync.Mutex
	collections map[string][]map[string]any
}

func newDataStore() *DataStore {
	return &DataStore{collections: map[string][]map[string]any{}}
}

// Push appends item to the named collection after assigning a UUID "id".
// Returns the stored item (with id set).
func (d *DataStore) Push(name string, item map[string]any) map[string]any {
	if item == nil {
		item = map[string]any{}
	}
	item["id"] = newID()
	d.mu.Lock()
	d.collections[name] = append(d.collections[name], item)
	d.mu.Unlock()
	return item
}

// List returns a shallow copy of all items in the named collection.
func (d *DataStore) List(name string) []map[string]any {
	d.mu.Lock()
	defer d.mu.Unlock()
	src := d.collections[name]
	out := make([]map[string]any, len(src))
	copy(out, src)
	return out
}

// Get returns the item with the given id from the named collection.
func (d *DataStore) Get(name, id string) (map[string]any, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, item := range d.collections[name] {
		if fmt.Sprint(item["id"]) == id {
			return item, true
		}
	}
	return nil, false
}

// Put replaces the item with the given id (upsert — inserts if not found).
// Returns true if an existing item was replaced, false if inserted.
func (d *DataStore) Put(name, id string, item map[string]any) bool {
	item["id"] = id
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, existing := range d.collections[name] {
		if fmt.Sprint(existing["id"]) == id {
			d.collections[name][i] = item
			return true
		}
	}
	d.collections[name] = append(d.collections[name], item)
	return false
}

// Patch merges partial fields into the item with the given id.
// The "id" field is protected and cannot be overwritten.
// Returns the updated item and true, or nil and false if not found.
func (d *DataStore) Patch(name, id string, partial map[string]any) (map[string]any, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, item := range d.collections[name] {
		if fmt.Sprint(item["id"]) == id {
			for k, v := range partial {
				if k == "id" {
					continue
				}
				item[k] = v
			}
			d.collections[name][i] = item
			return item, true
		}
	}
	return nil, false
}

// Delete removes the item with the given id. Returns true if found and removed.
func (d *DataStore) Delete(name, id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, item := range d.collections[name] {
		if fmt.Sprint(item["id"]) == id {
			d.collections[name] = append(d.collections[name][:i], d.collections[name][i+1:]...)
			return true
		}
	}
	return false
}

// Clear removes all items from the named collection.
func (d *DataStore) Clear(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.collections[name] = nil
}

// ClearAll removes all collections.
func (d *DataStore) ClearAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.collections = map[string][]map[string]any{}
}

// Names returns the names of all collections that have been written to.
func (d *DataStore) Names() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	names := make([]string, 0, len(d.collections))
	for name := range d.collections {
		names = append(names, name)
	}
	return names
}

// SetCollection replaces the entire named collection.
func (d *DataStore) SetCollection(name string, items []map[string]any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.collections[name] = items
}
