CREATE TABLE IF NOT EXISTS `games` (
  `id` int(10) unsigned NOT NULL,
  `name` varchar(30) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `games_addons` (
  `id` int(10) unsigned NOT NULL,
  `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `url` varchar(255) NOT NULL,
  `game_id` int(10) unsigned NOT NULL,
  `show_on_main` tinyint(3) unsigned NOT NULL,
  `ext_id` int(10) unsigned NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `games_addons_vars` (
  `addon_id` int(10) unsigned NOT NULL,
  `var_id` int(10) unsigned NOT NULL,
  `is_version` tinyint(3) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `games_mods` (
  `id` int(10) unsigned NOT NULL,
  `game_id` int(10) unsigned NOT NULL,
  `modstring` varchar(255) COLLATE utf8_bin NOT NULL,
  `description` varchar(255) COLLATE utf8_bin NOT NULL,
  `url` varchar(255) COLLATE utf8_bin NOT NULL,
  `is_verified` tinyint(4) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COLLATE=utf8_bin;

CREATE TABLE IF NOT EXISTS `games_vars` (
  `id` int(10) unsigned NOT NULL,
  `game_id` int(10) unsigned NOT NULL,
  `name` varchar(65) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `games_vars_values` (
  `id` int(10) unsigned NOT NULL,
  `variable_id` int(10) unsigned NOT NULL,
  `value` varbinary(255) NOT NULL,
  `first_known` int(10) unsigned NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `mode_types` (
`mode_id` int(10) unsigned NOT NULL,
  `name` varchar(16) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `object_types` (
  `id` int(10) unsigned NOT NULL,
  `name` varchar(30) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `os_types` (
`os_id` int(10) unsigned NOT NULL,
  `name` varchar(16) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `page_views` (
  `mod_id` int(10) unsigned NOT NULL,
  `addon_id` int(10) unsigned NOT NULL,
  `site_id` int(10) unsigned NOT NULL,
  `views` int(10) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `server_types` (
  `id` int(10) unsigned NOT NULL,
  `os` varchar(12) NOT NULL,
  `type` varchar(12) NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `stats_games` (
  `id` int(10) unsigned NOT NULL,
  `stamp` int(10) unsigned NOT NULL,
  `game_id` int(10) unsigned NOT NULL,
  `alive_count` mediumint(8) unsigned NOT NULL,
  `dead_count` mediumint(8) unsigned NOT NULL,
  `max_players` mediumint(8) unsigned NOT NULL,
  `total_players` mediumint(8) unsigned NOT NULL,
  `total_bots` mediumint(8) unsigned NOT NULL,
  `linux_servers` mediumint(8) unsigned NOT NULL,
  `windows_servers` mediumint(8) unsigned NOT NULL,
  `listen_servers` mediumint(8) unsigned NOT NULL
) ENGINE=InnoDB  DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `stats_games_addons` (
  `stats_id` int(10) unsigned NOT NULL,
  `object_id` int(10) unsigned NOT NULL,
  `server_count` mediumint(8) unsigned NOT NULL,
  `max_players` mediumint(8) unsigned NOT NULL,
  `total_players` mediumint(8) unsigned NOT NULL,
  `total_bots` mediumint(8) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `stats_games_values` (
  `stats_id` int(10) unsigned NOT NULL,
  `object_id` int(10) unsigned NOT NULL,
  `server_count` mediumint(8) unsigned NOT NULL,
  `max_players` mediumint(8) unsigned NOT NULL,
  `total_players` mediumint(8) unsigned NOT NULL,
  `total_bots` mediumint(8) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `stats_mods` (
  `stats_id` int(10) unsigned NOT NULL,
  `mod_id` int(10) unsigned NOT NULL,
  `server_type` tinyint(3) unsigned NOT NULL,
  `server_count` mediumint(8) unsigned NOT NULL,
  `max_players` mediumint(8) unsigned NOT NULL,
  `total_players` mediumint(8) unsigned NOT NULL,
  `total_bots` mediumint(8) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `stats_mods_addons` (
  `stats_id` int(10) unsigned NOT NULL,
  `object_id` int(10) unsigned NOT NULL,
  `mod_id` int(10) unsigned NOT NULL,
  `server_type` tinyint(3) unsigned NOT NULL,
  `server_count` mediumint(8) unsigned NOT NULL,
  `max_players` mediumint(8) unsigned NOT NULL,
  `total_players` mediumint(8) unsigned NOT NULL,
  `total_bots` mediumint(8) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `stats_mods_values` (
  `stats_id` int(10) unsigned NOT NULL,
  `object_id` int(10) unsigned NOT NULL,
  `mod_id` int(10) unsigned NOT NULL,
  `server_type` tinyint(3) unsigned NOT NULL,
  `server_count` mediumint(8) unsigned NOT NULL,
  `max_players` mediumint(8) unsigned NOT NULL,
  `total_players` mediumint(8) unsigned NOT NULL,
  `total_bots` mediumint(8) unsigned NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

ALTER TABLE `games`
 ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `name` (`name`);

ALTER TABLE `games_addons`
 ADD PRIMARY KEY (`id`);

ALTER TABLE `games_addons_vars`
 ADD PRIMARY KEY (`addon_id`,`var_id`);

ALTER TABLE `games_mods`
 ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `game_id` (`game_id`,`modstring`);

ALTER TABLE `games_vars`
 ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `game_id` (`game_id`,`name`);

ALTER TABLE `games_vars_values`
 ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `variable_id` (`variable_id`,`value`);

ALTER TABLE `mode_types`
 ADD PRIMARY KEY (`mode_id`), ADD KEY `name` (`name`);

ALTER TABLE `object_types`
 ADD PRIMARY KEY (`id`);

ALTER TABLE `os_types`
 ADD PRIMARY KEY (`os_id`), ADD KEY `name` (`name`);

ALTER TABLE `page_views`
 ADD PRIMARY KEY (`mod_id`,`addon_id`,`site_id`);

ALTER TABLE `server_types`
 ADD PRIMARY KEY (`id`);

ALTER TABLE `stats_games`
 ADD PRIMARY KEY (`id`);

ALTER TABLE `stats_mods`
 ADD PRIMARY KEY (`stats_id`,`mod_id`,`server_type`);
