package database

import "testing"

func TestUserIsAdminRoundTrip(t *testing.T) {
	db, err := Open("file:booltest?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)

	if err := repo.User.Create("admin", "h1", true); err != nil {
		t.Fatal(err)
	}
	if err := repo.User.Create("plain", "h2", false); err != nil {
		t.Fatal(err)
	}

	admin, err := repo.User.GetByUsername("admin")
	if err != nil {
		t.Fatal(err)
	}
	if !admin.IsAdmin {
		t.Fatalf("admin.IsAdmin = false, want true")
	}

	plain, err := repo.User.GetByUsername("plain")
	if err != nil {
		t.Fatal(err)
	}
	if plain.IsAdmin {
		t.Fatalf("plain.IsAdmin = true, want false")
	}

	// AdminExists relies on is_admin = 1 matching a bool-bound insert.
	exists, err := repo.User.AdminExists()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("AdminExists() = false, want true")
	}
}
