// vim: set ts=4 sw=4 tw=99 noet:
//
// Blaster (C) Copyright 2014 AlliedModders LLC
// Licensed under the GNU General Public License, version 3 or higher.
// See LICENSE.txt for more details.
package main

import (
	"time"
)

type Stats struct {
	BaseStat
	ServerCount int64
}

type StatsKey struct {
	ModId   int64
	Type    int64
	AddonId int64
	ValueId int64
}

func (this *StatsKey) TableName() string {
	switch {
	case this.AddonId != 0:
		if this.ModId == 0 {
			return "stats_games_addons"
		}
		return "stats_mods_addons"
	case this.ValueId != 0:
		if this.ModId == 0 {
			return "stats_games_values"
		}
		return "stats_mods_values"
	default:
		return "stats_mods"
	}
}

type ValueKey struct {
	VariableId int64
	Value      string
}

type StatsCollector struct {
	db        *Database
	game_id   int64
	global    *GameStat
	rows      map[StatsKey]*Stats
	mods      map[string]*GameMod
	addonVars map[string]*GameAddonVar
	values    map[ValueKey]*GameVarValue
}

func NewStatsCollector(db *Database, game_id int64) *StatsCollector {
	modMap := map[string]*GameMod{}
	for _, mod := range getGameMods(db, game_id) {
		modMap[mod.ModString] = mod
	}

	valueMap := map[ValueKey]*GameVarValue{}
	for _, gameVar := range getGameVarValues(db, game_id) {
		valueMap[ValueKey{gameVar.VariableId, gameVar.Value}] = gameVar
	}

	addonVarMap := map[string]*GameAddonVar{}
	for _, addonVar := range getAddonVars(db, game_id) {
		addonVarMap[addonVar.Name] = addonVar
	}

	return &StatsCollector{
		db:      db,
		game_id: game_id,
		global: &GameStat{
			GameId: game_id,
		},
		rows:      map[StatsKey]*Stats{},
		mods:      modMap,
		addonVars: addonVarMap,
		values:    valueMap,
	}
}

func (this *StatsCollector) get(key StatsKey) *Stats {
	stats, ok := this.rows[key]
	if ok {
		return stats
	}
	stats = &Stats{}
	this.rows[key] = stats
	return stats
}

func (this *StatsCollector) getValue(varId int64, value string) int64 {
	key := ValueKey{varId, value}
	obj, ok := this.values[key]
	if ok {
		return obj.Id
	}

	obj = &GameVarValue{
		VariableId: varId,
		Value:      value,
		FirstKnown: time.Now().Unix(),
	}
	this.db.Insert(obj)

	this.values[key] = obj
	return obj.Id
}

func (this *StatsCollector) getAddonByVarName(name string) *GameAddonVar {
	addonVar, ok := this.addonVars[name]
	if !ok {
		return nil
	}
	return addonVar
}

func (this *StatsCollector) getMod(server *Server) int64 {
	mod, ok := this.mods[server.Info.Folder]
	if ok {
		return mod.Id
	}

	mod = &GameMod{
		GameId:      this.game_id,
		ModString:   server.Info.Folder,
		Description: server.Info.Game,
	}
	this.db.Insert(mod)

	this.mods[mod.ModString] = mod
	return mod.Id
}

func (this *StatsCollector) finish() {
	// Use idle conns now.
	this.db.conn.SetMaxIdleConns(1)

	// Insert the global row.
	this.db.Insert(this.global)

	// Insert all the stats rows.
	for key, stat := range this.rows {
		switch key.TableName() {
		case "stats_mods":
			row := &GameModStat{
				BaseStat:    stat.BaseStat,
				StatsId:     this.global.Id,
				ModId:       key.ModId,
				ServerType:  key.Type,
				ServerCount: stat.ServerCount,
			}
			this.db.Insert(row)
		case "stats_games_addons":
			row := &GameAddonStat{
				BaseStat:    stat.BaseStat,
				StatsId:     this.global.Id,
				ServerCount: stat.ServerCount,
				ObjectId:    key.AddonId,
			}
			this.db.Insert(row)
		case "stats_games_values":
			row := &GameValueStat{
				BaseStat:    stat.BaseStat,
				StatsId:     this.global.Id,
				ServerCount: stat.ServerCount,
				ObjectId:    key.ValueId,
			}
			this.db.Insert(row)
		case "stats_mods_addons":
			row := &GameModAddonStat{
				GameModStat: GameModStat{
					BaseStat:    stat.BaseStat,
					StatsId:     this.global.Id,
					ServerCount: stat.ServerCount,
					ModId:       key.ModId,
					ServerType:  key.Type,
				},
				ObjectId: key.AddonId,
			}
			this.db.Insert(row)
		case "stats_mods_values":
			row := &GameModValueStat{
				GameModStat: GameModStat{
					BaseStat:    stat.BaseStat,
					StatsId:     this.global.Id,
					ServerCount: stat.ServerCount,
					ModId:       key.ModId,
					ServerType:  key.Type,
				},
				ObjectId: key.ValueId,
			}
			this.db.Insert(row)
		}
	}

	this.db.Exec("UPDATE stats_games SET stamp = UNIX_TIMESTAMP() WHERE id = ?", this.global.Id)
}
