// Package domain holds cross-service primitives shared by every service:
// roles and the trusted identity headers the gateway forwards downstream.
package domain

// Role identifies a portal user's role.
type Role string

const (
	RoleOfficer Role = "officer"
	RoleMember  Role = "member"
)

// Valid reports whether r is a recognized role.
func (r Role) Valid() bool {
	return r == RoleOfficer || r == RoleMember
}

// String returns the role as a plain string.
func (r Role) String() string { return string(r) }

// Trusted identity headers. The gateway sets these after verifying the JWT;
// downstream services read them and must only be reachable via the gateway.
const (
	HeaderUserID    = "X-User-Id"
	HeaderUserRole  = "X-User-Role"
	HeaderUserEmail = "X-User-Email"
	HeaderUserName  = "X-User-Name"
)

// ViolationType is a parking violation category. Its string value is the key
// used for base amounts in a fine ruleset.
type ViolationType string

const (
	ViolationExpiredMeter    ViolationType = "expired_meter"
	ViolationNoParkingZone   ViolationType = "no_parking_zone"
	ViolationBlockingHydrant ViolationType = "blocking_hydrant"
	ViolationDisabledSpot    ViolationType = "disabled_spot"
)

// ViolationTypes returns every known violation type, in display order.
func ViolationTypes() []ViolationType {
	return []ViolationType{
		ViolationExpiredMeter,
		ViolationNoParkingZone,
		ViolationBlockingHydrant,
		ViolationDisabledSpot,
	}
}

// Valid reports whether v is a recognized violation type.
func (v ViolationType) Valid() bool {
	switch v {
	case ViolationExpiredMeter, ViolationNoParkingZone, ViolationBlockingHydrant, ViolationDisabledSpot:
		return true
	default:
		return false
	}
}

// String returns the violation type as a plain string.
func (v ViolationType) String() string { return string(v) }
