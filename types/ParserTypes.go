package types

type ElementInfo struct {
	Name string
	ID string
}

func NewElementInfo(n string, id string) *ElementInfo {
	return &ElementInfo{Name: n, ID: id}
}

// TODO: Add more functions to support any type if necessary
