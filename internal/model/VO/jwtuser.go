package VO

type JwtUser struct {
	Id       int64
	Username string
	Status   int
	Roles    []string
}
