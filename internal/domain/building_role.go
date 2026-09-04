package domain

type RoleName string

const (
	roleNameTenant RoleName = "tenant"
	roleNameOwner  RoleName = "owner"
	roleNameAdmin  RoleName = "admin"
)

// Access levels — authorization always compares against Level, never Code,
// so a new role can be inserted between existing ones without code changes.
const (
	levelTenant int16 = 20
	levelOwner  int16 = 40
	levelAdmin  int16 = 100
)

type Role struct {
	ID    int16
	Name  RoleName
	Level int16
}

var (
	RoleNone   = Role{ID: 0, Name: "", Level: -1}
	RoleTenant = Role{ID: 1, Name: roleNameTenant, Level: levelTenant}
	RoleOwner  = Role{ID: 2, Name: roleNameOwner, Level: levelOwner}
	RoleAdmin  = Role{ID: 3, Name: roleNameAdmin, Level: levelAdmin}
)

func GetRoleByName(name RoleName) Role {
	switch name {
	case roleNameTenant:
		return RoleTenant
	case roleNameOwner:
		return RoleOwner
	case roleNameAdmin:
		return RoleAdmin
	default:
		return RoleNone
	}
}
