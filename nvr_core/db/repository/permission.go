package repository

import (
    "context"
    "database/sql"
    "fmt"

    "nvr_core/db/models"
)

type PermissionRepository interface {
    // GetUserPermissionCodes returns a flat array of strings like ["camera:view", "timeline:play"]
    GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error)

    // Standard CRUD for administrative UI
    GetAll(ctx context.Context) ([]*models.Permission, error)
    GetRolePermissions(ctx context.Context, roleID int64) ([]*models.Permission, error)
}

type permissionRepo struct {
    db *sql.DB
}

func NewPermissionRepository(db *sql.DB) PermissionRepository {
    return &permissionRepo{db: db}
}

// GetUserPermissionCodes executes the UNION query to gather all access rights for a user.
func (r *permissionRepo) GetUserPermissionCodes(ctx context.Context, userID int64) ([]string, error) {
    // Part 1: Permissions from the Role
    // Part 2: Permissions directly granted to the User (Exceptions)
    query := `
        SELECT p.code 
        FROM users u
        JOIN role_permissions rp ON u.role_id = rp.role_id
        JOIN permissions p ON rp.permission_id = p.id
        WHERE u.id = ? AND u.is_active = 1

        UNION

        SELECT p.code 
        FROM users u
        JOIN user_permissions up ON u.id = up.user_id
        JOIN permissions p ON up.permission_id = p.id
        WHERE u.id = ? AND u.is_active = 1;
    `

    // Note: We pass userID twice to satisfy both '?' placeholders
    rows, err := r.db.QueryContext(ctx, query, userID, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user permissions: %w", err)
    }
    defer rows.Close()

    var codes []string
    for rows.Next() {
        var code string
        if err := rows.Scan(&code); err != nil {
            return nil, err
        }
        codes = append(codes, code)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return codes, nil
}

// GetAll fetches all available permissions to populate the Admin UI checkboxes.
func (r *permissionRepo) GetAll(ctx context.Context) ([]*models.Permission, error) {
    query := `SELECT id, code, description FROM permissions ORDER BY code ASC`

    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var perms []*models.Permission
    for rows.Next() {
        var p models.Permission
        if err := rows.Scan(&p.ID, &p.Code, &p.Description); err != nil {
            return nil, err
        }
        perms = append(perms, &p)
    }
    return perms, rows.Err()
}

// GetRolePermissions fetches only the permissions assigned to a specific role.
func (r *permissionRepo) GetRolePermissions(ctx context.Context, roleID int64) ([]*models.Permission, error) {
    query := `
        SELECT p.id, p.code, p.description 
        FROM permissions p
        JOIN role_permissions rp ON p.id = rp.permission_id
        WHERE rp.role_id = ?
        ORDER BY p.code ASC
    `

    rows, err := r.db.QueryContext(ctx, query, roleID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var perms []*models.Permission
    for rows.Next() {
        var p models.Permission
        if err := rows.Scan(&p.ID, &p.Code, &p.Description); err != nil {
            return nil, err
        }
        perms = append(perms, &p)
    }
    return perms, rows.Err()
}