INSERT INTO `games` (`id`, `name`) VALUES
(1, 'Half-Life 1'),
(2, 'Half-Life 2');

--
-- Dumping data for table `mode_types`
--

INSERT INTO `mode_types` (`mode_id`, `name`) VALUES
(1, 'dedicated'),
(2, 'listen'),
(3, 'hltv');

--
-- Dumping data for table `object_types`
--

INSERT INTO `object_types` (`id`, `name`) VALUES
(1, 'addon'),
(2, 'value');

--
-- Dumping data for table `os_types`
--

INSERT INTO `os_types` (`os_id`, `name`) VALUES
(1, 'windows'),
(2, 'linux'),
(3, 'mac');

--
-- Dumping data for table `server_types`
--

INSERT INTO `server_types` (`id`, `os`, `type`) VALUES
(1, 'linux', 'dedicated'),
(2, 'windows', 'dedicated'),
(3, 'windows', 'listen'),
(4, 'mac', 'dedicated');
