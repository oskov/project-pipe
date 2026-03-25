// Package sample is a fixture used by GoParseService tests.
// It contains one of every top-level definition kind.
package sample

import "fmt"

// MaxItems is the upper limit for item collections.
const MaxItems = 100

// Sentinel errors used across the package.
const (
	ErrNotFound = "not found"
	ErrInvalid  = "invalid"
)

// DefaultTimeout holds the default operation timeout in seconds.
var DefaultTimeout = 30

// Counters tracks various runtime statistics.
var (
	RequestCount int
	ErrorCount   int
)

// Item represents a single work item.
type Item struct {
	ID    string
	Value string
}

// Repository defines persistence operations for items.
type Repository interface {
	Get(id string) (*Item, error)
	Save(item *Item) error
}

// mapStore is an alias for a plain map used as storage.
type mapStore = map[string]*Item

// NewItem creates a new Item with the given id and value.
func NewItem(id, value string) *Item {
	return &Item{ID: id, Value: value}
}

// String returns a human-readable representation of the item.
func (i *Item) String() string {
	return fmt.Sprintf("Item(%s=%s)", i.ID, i.Value)
}

// Validate checks that the item fields are non-empty.
func (i *Item) Validate() error {
	if i.ID == "" || i.Value == "" {
		return fmt.Errorf("%s: id and value are required", ErrInvalid)
	}
	return nil
}
