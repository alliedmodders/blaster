package main

import (
	"database/sql"
	"fmt"

	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kylelemons/go-gypsy/yaml"
)

type Database struct {
	*gorp.DbMap
	conn *sql.DB
}

func getYamlKey(config *yaml.File, keyPrefix string, key string) string {
	result, err := config.Get(fmt.Sprintf("%s.%s", keyPrefix, key))
	if err != nil {
		panic(err)
	}
	return result
}

func getDsn(config *yaml.File, key string) string {
	host := getYamlKey(config, key, "host")
	username := getYamlKey(config, key, "username")
	dbname := getYamlKey(config, key, "dbname")
	password, _ := config.Get(fmt.Sprintf("%s.%s", key, "password"))

	if password == "" {
		return fmt.Sprintf("%s@tcp(%s)/%s", username, host, dbname)
	}
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, dbname)
}

func getDatabase(config *yaml.File, key string) *Database {
	dsn := getDsn(config, key)
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	db := &Database{
		&gorp.DbMap{
			Db: conn,
			Dialect: gorp.MySQLDialect{
				Engine:   "MyISAM",
				Encoding: "UTF8",
			},
		},
		conn,
	}

	db.AddTableWithName(Game{}, "games").
		SetKeys(true, "id")
	db.AddTableWithName(GameMod{}, "games_mods").
		SetKeys(true, "id")
	db.AddTableWithName(GameVarValue{}, "games_vars_values").
		SetKeys(true, "id")
	db.AddTableWithName(GameStat{}, "stats_games").
		SetKeys(true, "id")
	db.AddTableWithName(GameAddonStat{}, "stats_games_addons")
	db.AddTableWithName(GameValueStat{}, "stats_games_values")
	db.AddTableWithName(GameModStat{}, "stats_mods")
	db.AddTableWithName(GameModAddonStat{}, "stats_mods_addons")
	db.AddTableWithName(GameModValueStat{}, "stats_mods_values")

	return db
}

func (this *Database) Select(holder interface{}, query string, bindings ...interface{}) {
	_, err := this.DbMap.Select(holder, query, bindings...)
	if err != nil {
		panic(err)
	}
}

func (this *Database) Insert(holder interface{}) {
	if err := this.DbMap.Insert(holder); err != nil {
		panic(err)
	}
}

func (this *Database) Close() {
	this.conn.Close()
}
