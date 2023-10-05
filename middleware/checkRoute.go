package middleware

import "walk-server/model"

func CheckRoute(admin *model.Admin, team *model.Team) bool {
	if team.Point+1 == admin.Point && team.Route == admin.Route {
		return true
	} else if team.Point+1 != admin.Point {
		return false
	} else if (team.Route == 4 && admin.Route == 5) || (team.Route == 5 && admin.Route == 4) {
		return true
	} else if (team.Route == 2 && admin.Route == 3 && admin.Point < 5) || (team.Route == 3 && admin.Route == 2 && admin.Point < 5) {
		return true
	}
	return false
}
