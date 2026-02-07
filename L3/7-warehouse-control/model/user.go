package model

type RoleType uint8

const (
	Viewer RoleType = 1 << iota
	Owner
	Manager
	Admin
)

type User struct {
	ID       int64
	Role     RoleType
	Name     string
	Password string
}
