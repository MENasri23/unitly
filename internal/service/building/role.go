package building

type RoleCode string

const (
	roleCodeTenant RoleCode = "tenant"
	roleCodeOwner  RoleCode = "owner"
	roleCodeAdmin  RoleCode = "admin"
)

// Access levels — authorization always compares against Level, never Code,
// so a new role can be inserted between existing ones without code changes.
const (
	LevelTenant int16 = 20
	LevelOwner  int16 = 40
	LevelAdmin  int16 = 100
)

type Role struct {
	ID    int16
	Code  RoleCode
	Level int16
}

var (
	RoleTenant = Role{ID: 1, Code: roleCodeTenant, Level: LevelTenant}
	RoleOwner  = Role{ID: 2, Code: roleCodeOwner, Level: LevelOwner}
	RoleAdmin  = Role{ID: 3, Code: roleCodeAdmin, Level: LevelAdmin}
)