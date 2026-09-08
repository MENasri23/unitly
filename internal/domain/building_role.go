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
	RoleTenant = Role{ID: 1, Name: roleNameTenant, Level: levelTenant}
	RoleOwner  = Role{ID: 2, Name: roleNameOwner, Level: levelOwner}
	RoleAdmin  = Role{ID: 3, Name: roleNameAdmin, Level: levelAdmin}
)

func GetRoleByName(name RoleName) (Role, bool) {
	switch name {
	case roleNameTenant:
		return RoleTenant, true
	case roleNameOwner:
		return RoleOwner, true
	case roleNameAdmin:
		return RoleAdmin, true
	default:
		return Role{}, false
	}
}

func GetRoleByID(id int16) (Role, bool) {
	switch id {
	case RoleTenant.ID:
		return RoleTenant, true
	case RoleOwner.ID:
		return RoleOwner, true
	case RoleAdmin.ID:
		return RoleAdmin, true
	default:
		return Role{}, false
	}
}
