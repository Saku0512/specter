package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// DataStore is a named-collection in-memory CRUD store.
// Each collection holds objects identified by a string "id" field.
type DataStore struct {
	mu             sync.Mutex
	collections    map[string][]map[string]any
	persistPath    string
	lastPersistErr error
}

func newDataStore() *DataStore {
	return &DataStore{collections: map[string][]map[string]any{}}
}

func newDataStoreWithFile(path string) (*DataStore, error) {
	store := newDataStore()
	if path == "" {
		return store, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			store.persistPath = path
			return store, nil
		}
		return nil, err
	}
	if len(data) > 0 {
		var collections map[string][]map[string]any
		if err := json.Unmarshal(data, &collections); err != nil {
			return nil, err
		}
		store.collections = cloneCollections(collections)
	}
	store.persistPath = path
	return store, nil
}

func cloneCollections(collections map[string][]map[string]any) map[string][]map[string]any {
	copied := map[string][]map[string]any{}
	for name, items := range collections {
		copiedItems := make([]map[string]any, len(items))
		for i, item := range items {
			copiedItem := make(map[string]any, len(item))
			for k, v := range item {
				copiedItem[k] = v
			}
			copiedItems[i] = copiedItem
		}
		copied[name] = copiedItems
	}
	return copied
}

func (d *DataStore) snapshotLocked() map[string][]map[string]any {
	return cloneCollections(d.collections)
}

func (d *DataStore) persistSnapshot(collections map[string][]map[string]any) error {
	if d.persistPath == "" {
		return nil
	}
	data, err := json.MarshalIndent(collections, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(d.persistPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(d.persistPath)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, d.persistPath); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

func (d *DataStore) persist() error {
	d.mu.Lock()
	snapshot := d.snapshotLocked()
	d.mu.Unlock()
	return d.persistSnapshot(snapshot)
}

func (d *DataStore) saveAfterMutation() {
	err := d.persist()
	d.mu.Lock()
	d.lastPersistErr = err
	d.mu.Unlock()
}

func (d *DataStore) LastPersistError() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.lastPersistErr
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
	d.saveAfterMutation()
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
	for i, existing := range d.collections[name] {
		if fmt.Sprint(existing["id"]) == id {
			d.collections[name][i] = item
			d.mu.Unlock()
			d.saveAfterMutation()
			return true
		}
	}
	d.collections[name] = append(d.collections[name], item)
	d.mu.Unlock()
	d.saveAfterMutation()
	return false
}

// Patch merges partial fields into the item with the given id.
// The "id" field is protected and cannot be overwritten.
// Returns the updated item and true, or nil and false if not found.
func (d *DataStore) Patch(name, id string, partial map[string]any) (map[string]any, bool) {
	d.mu.Lock()
	for i, item := range d.collections[name] {
		if fmt.Sprint(item["id"]) == id {
			for k, v := range partial {
				if k == "id" {
					continue
				}
				item[k] = v
			}
			d.collections[name][i] = item
			d.mu.Unlock()
			d.saveAfterMutation()
			return item, true
		}
	}
	d.mu.Unlock()
	return nil, false
}

// Delete removes the item with the given id. Returns true if found and removed.
func (d *DataStore) Delete(name, id string) bool {
	d.mu.Lock()
	for i, item := range d.collections[name] {
		if fmt.Sprint(item["id"]) == id {
			d.collections[name] = append(d.collections[name][:i], d.collections[name][i+1:]...)
			d.mu.Unlock()
			d.saveAfterMutation()
			return true
		}
	}
	d.mu.Unlock()
	return false
}

// Clear removes all items from the named collection.
func (d *DataStore) Clear(name string) {
	d.mu.Lock()
	d.collections[name] = nil
	d.mu.Unlock()
	d.saveAfterMutation()
}

// ClearAll removes all collections.
func (d *DataStore) ClearAll() {
	d.mu.Lock()
	d.collections = map[string][]map[string]any{}
	d.mu.Unlock()
	d.saveAfterMutation()
}

// ReplaceAll replaces every collection with a shallow copy of collections.
func (d *DataStore) ReplaceAll(collections map[string][]map[string]any) {
	d.mu.Lock()
	d.collections = cloneCollections(collections)
	d.mu.Unlock()
	d.saveAfterMutation()
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
	d.collections[name] = items
	d.mu.Unlock()
	d.saveAfterMutation()
}
