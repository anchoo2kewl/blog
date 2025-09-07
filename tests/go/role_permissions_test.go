package gotests

import (
    "testing"
    m "anshumanbiswas.com/blog/models"
)

func TestIsAdmin(t *testing.T) {
    if !m.IsAdmin(m.RoleAdministrator) {
        t.Fatalf("expected admin role to be admin")
    }
    if m.IsAdmin(m.RoleCommenter) || m.IsAdmin(m.RoleEditor) || m.IsAdmin(m.RoleViewer) {
        t.Fatalf("only administrator should be admin")
    }
}

func TestPermissionsMatrix(t *testing.T) {
    // Commenter
    p := m.GetPermissions(m.RoleCommenter)
    if p.CanEditPosts || p.CanManageUsers || p.CanViewUnpublished {
        t.Fatalf("commenter permissions too broad: %+v", p)
    }
    // Editor
    p = m.GetPermissions(m.RoleEditor)
    if !p.CanEditPosts || p.CanManageUsers {
        t.Fatalf("editor permissions incorrect: %+v", p)
    }
    // Admin
    p = m.GetPermissions(m.RoleAdministrator)
    if !(p.CanEditPosts && p.CanManageUsers && p.CanViewUnpublished) {
        t.Fatalf("admin permissions incorrect: %+v", p)
    }
}

