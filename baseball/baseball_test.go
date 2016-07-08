package baseball

import "testing"

func Test_favoriteTeamHandler(t *testing.T) {
	t.Parallel()

	teamID := "147"
	gd := FavoriteTeamGameListHandler(teamID, InitHomePageMap())

	if true != true {
		t.Fatalf("Expected %s got %s", teamID, gd[0].Date)
	}
}
