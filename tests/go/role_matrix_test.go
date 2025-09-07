package gotests

import (
    "testing"
    m "anshumanbiswas.com/blog/models"
)

func TestRoleMatrix(t *testing.T) {
    // Commenter
    c := m.GetPermissions(m.RoleCommenter)
    if !c.CanComment || c.CanEditPosts || c.CanManageUsers {
        t.Fatalf("commenter perms incorrect: %+v", c)
    }
    // Viewer
    v := m.GetPermissions(m.RoleViewer)
    if !v.CanComment || v.CanEditPosts || v.CanManageUsers {
        t.Fatalf("viewer perms incorrect: %+v", v)
    }
    // Editor
    e := m.GetPermissions(m.RoleEditor)
    if !e.CanEditPosts || e.CanManageUsers {
        t.Fatalf("editor perms incorrect: %+v", e)
    }
    // Admin
    a := m.GetPermissions(m.RoleAdministrator)
    if !(a.CanComment && a.CanEditPosts && a.CanManageUsers && a.CanViewUnpublished) {
        t.Fatalf("admin perms incorrect: %+v", a)
    }
}

