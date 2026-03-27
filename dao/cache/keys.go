package cache

import "fmt"

func UserProfileKey(openID string) string {
	return fmt.Sprintf("walk:user:profile:%s", openID)
}

func TeamInfoKey(teamID int64) string {
	return fmt.Sprintf("walk:user:team:info:%d", teamID)
}

func TeamNameExistsKey(name string) string {
	return fmt.Sprintf("walk:user:team:name:exists:%s", name)
}

func TeamCreateLockKey(openID string) string {
	return fmt.Sprintf("walk:user:team:create:lock:%s", openID)
}

func TeamJoinLockKey(openID string) string {
	return fmt.Sprintf("walk:user:team:join:lock:%s", openID)
}

func TeamChangeLockKey(teamID int64) string {
	return fmt.Sprintf("walk:user:team:change:lock:%d", teamID)
}

func WechatLoginCodeKey(code string) string {
	return fmt.Sprintf("walk:user:wechat:login:code:%s", code)
}

func RateRegisterKey(openID string) string {
	return fmt.Sprintf("walk:user:rate:register:%s", openID)
}

func RateTeamKey(openID string) string {
	return fmt.Sprintf("walk:user:rate:team:%s", openID)
}

func AdminUserInfoKey(userID int64) string {
	return fmt.Sprintf("walk:admin:user:info:%d", userID)
}

func AdminTeamsKey() string {
	return "walk:admin:teams"
}

func AdminUserStatusKey(userID int64) string {
	return fmt.Sprintf("walk:admin:user:status:%d", userID)
}
