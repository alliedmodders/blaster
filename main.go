package main

import (
	"fmt"

	. "github.com/alliedmodders/blaster/valve"
)

func main() {
	master := NewMasterServerQuerier("hl1master.steampowered.com:27011")
	master.FilterAppIds(HL1Apps)
	batch := 0
	err := master.Query(func(servers ServerList) error {
		for _, server := range servers {
			fmt.Printf("[%d] %s\n", batch, server)
		}
		batch++
		return nil
	})
	if err != nil {
		panic(err)
	}
}
