package middleware

import "walk-server/model"

func CheckRoute(admin *model.Admin, team *model.Team) bool {
	if team.Point+1 == admin.Point && team.Route == admin.Route {
		return true
	} else if team.Point+1 != admin.Point {
		return false
	} else if (team.Route == 4 && admin.Route == 5) || (team.Route == 5 && admin.Route == 4) {
		return true
	}
	return false
}
