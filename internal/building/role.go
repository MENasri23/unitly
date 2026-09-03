package building

type RoleCode string

const (
	roleNameTenant RoleCode = "tenant"
	roleNameOwner  RoleCode = "owner"
	roleNameAdmin  RoleCode = "admin"
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
	Name  RoleCode
	Level int16
}

var (
	RoleTenant = Role{ID: 1, Name: roleNameTenant, Level: levelTenant}
	RoleOwner  = Role{ID: 2, Name: roleNameOwner, Level: levelOwner}
	RoleAdmin  = Role{ID: 3, Name: roleNameAdmin, Level: levelAdmin}
)
