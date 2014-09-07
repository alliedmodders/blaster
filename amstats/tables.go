package main

type Game struct {
	Id   int64  `db:"id"`
	Name string `db:"name"`
}

type GameMod struct {
	Id          int64  `db:"id"`
	GameId      int64  `db:"game_id"`
	ModString   string `db:"modstring"`
	Description string `db:"description"`
	Url         string `db:"url"`
	IsVerified  int    `db:"is_verified"`
}

type GameAddonVar struct {
	VariableId int64  `db:"var_id"`
	AddonId    int64  `db:"addon_id"`
	Name       string `db:"name"`
}

type GameVarValue struct {
	Id         int64  `db:"id"`
	VariableId int64  `db:"variable_id"`
	Value      string `db:"value"`
	FirstKnown int64  `db:"first_known"`
}

type BaseStat struct {
	MaxPlayers   int64 `db:"max_players"`
	TotalPlayers int64 `db:"total_players"`
	TotalBots    int64 `db:"total_bots"`
}

type GameStat struct {
	BaseStat
	Id             int64 `db:"id"`
	Stamp          int64 `db:"stamp"`
	GameId         int64 `db:"game_id"`
	AliveCount     int64 `db:"alive_count"`
	DeadCount      int64 `db:"dead_count"`
	LinuxServers   int64 `db:"linux_servers"`
	WindowsServers int64 `db:"windows_servers"`
	ListenServers  int64 `db:"listen_servers"`
}

type GameObjectStat struct {
	BaseStat
	StatsId     int64 `db:"stats_id"`
	ObjectId    int64 `db:"object_id"`
	ServerCount int64 `db:"server_count"`
}
type GameAddonStat GameObjectStat
type GameValueStat GameObjectStat

type GameModStat struct {
	BaseStat
	StatsId     int64 `db:"stats_id"`
	ModId       int64 `db:"mod_id"`
	ServerType  int64 `db:"server_type"`
	ServerCount int64 `db:"server_count"`
}

type GameModObjectStat struct {
	GameModStat
	ObjectId int64 `db:"object_id"`
}
type GameModAddonStat GameModObjectStat
type GameModValueStat GameModObjectStat

func getAddonVars(db *Database, game_id int64) []*GameAddonVar {
	query := `
		SELECT gav.addon_id, gav.var_id, gv.name
		FROM games_addons_vars gav
		JOIN games_vars gv
			ON gav.var_id = gv.id
		JOIN games_addons ga
			ON gav.addon_id = ga.id
		WHERE ga.game_id = ?
	`

	var vars []*GameAddonVar
	db.Select(&vars, query, game_id)
	return vars
}

func getGameMods(db *Database, game_id int64) []*GameMod {
	var mods []*GameMod
	db.Select(&mods, "SELECT * from games_mods WHERE game_id = ?", game_id)
	return mods
}

func getGameVarValues(db *Database, game_id int64) []*GameVarValue {
	query := `
		SELECT gvv.id, gvv.variable_id, gvv.value, gvv.first_known
		FROM games_vars_values gvv
		JOIN games_vars gv
			ON gvv.variable_id = gv.id
		WHERE gv.game_id = ?
	`
	var values []*GameVarValue
	db.Select(&values, query, game_id)
	return values
}
