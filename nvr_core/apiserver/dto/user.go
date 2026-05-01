package dto

type UpdatePermissionsRequest struct {
	PermissionIDs []int64 `json:"permission_ids"`
}