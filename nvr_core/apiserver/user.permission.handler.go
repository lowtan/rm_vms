package apiserver

import (
	"encoding/json"
	"net/http"
	"nvr_core/apiserver/dto"
	"strconv"
)

// HandleUpdateUserPermissions expects: PUT /api/users/{id}/permissions
func (api *APIServer) HandleUpdateUserPermissions(w http.ResponseWriter, r *http.Request) {
	// In a real scenario, you extract the Admin's ID from the JWT middleware context here.
	adminID := int64(1) 

	targetIDStr := r.PathValue("id")
	targetUserID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req dto.UpdatePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Call the service layer to perform the transactional swap
	err = api.Services.User.UpdateUserPermissions(r.Context(), adminID, targetUserID, req.PermissionIDs)
	if err != nil {
		// Log error internally, return generic 500 or 404 to client
		http.Error(w, "Failed to update permissions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success"}`))
}