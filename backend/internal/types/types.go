package types

type UserRole string

const (
	UR_User  UserRole = "user"
	UR_Admin UserRole = "admin"
)

type TokenType string

const (
	TT_Access  TokenType = "access"
	TT_Refresh TokenType = "refresh"
)

type MetadataKey string

const (
	MD_Authorization MetadataKey = "authorization"
	MD_UserID        MetadataKey = "user_id"
	MD_UserRole      MetadataKey = "user_role"
)
