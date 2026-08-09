jaisheesh@LAPTOP-2U6VV3L0:/mnt/c/Users/Balagowni Jaisheesh/Desktop/Projects/MCP-Go$ docker compose ps
NAME             IMAGE       COMMAND                  SERVICE   CREATED          STATUS                    PORTS
mcp-go-mysql-1   mysql:8.0   "docker-entrypoint.s…"   mysql     11 minutes ago   Up 11 minutes (healthy)   3306/tcp
jaisheesh@LAPTOP-2U6VV3L0:/mnt/c/Users/Balagowni Jaisheesh/Desktop/Projects/MCP-Go$ docker compose logs mysql --no-log-prefix --tail=200
2026-08-08 18:02:30+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 8.0.46-1.el9 started.
2026-08-08 18:02:31+00:00 [Note] [Entrypoint]: Switching to dedicated user 'mysql'
2026-08-08 18:02:31+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 8.0.46-1.el9 started.
2026-08-08 18:02:31+00:00 [Note] [Entrypoint]: Initializing database files
2026-08-08T18:02:31.337759Z 0 [Warning] [MY-011068] [Server] The syntax '--skip-host-cache' is deprecated and will be removed in a future release. Please use SET GLOBAL host_cache_size=0 instead.
2026-08-08T18:02:31.337950Z 0 [System] [MY-013169] [Server] /usr/sbin/mysqld (mysqld 8.0.46) initializing of server in progress as process 80
2026-08-08T18:02:31.354752Z 1 [System] [MY-013576] [InnoDB] InnoDB initialization has started.
2026-08-08T18:02:32.439296Z 1 [System] [MY-013577] [InnoDB] InnoDB initialization has ended.
2026-08-08T18:02:35.068272Z 6 [Warning] [MY-010453] [Server] root@localhost is created with an empty password ! Please consider switching off the --initialize-insecure option.
2026-08-08 18:02:39+00:00 [Note] [Entrypoint]: Database files initialized
2026-08-08 18:02:39+00:00 [Note] [Entrypoint]: Starting temporary server
2026-08-08T18:02:40.305938Z 0 [Warning] [MY-011068] [Server] The syntax '--skip-host-cache' is deprecated and will be removed in a future release. Please use SET GLOBAL host_cache_size=0 instead.
2026-08-08T18:02:40.308886Z 0 [System] [MY-010116] [Server] /usr/sbin/mysqld (mysqld 8.0.46) starting as process 147
2026-08-08T18:02:40.326978Z 1 [System] [MY-013576] [InnoDB] InnoDB initialization has started.
2026-08-08T18:02:40.633411Z 1 [System] [MY-013577] [InnoDB] InnoDB initialization has ended.
2026-08-08T18:02:41.087990Z 0 [Warning] [MY-010068] [Server] CA certificate ca.pem is self signed.
2026-08-08T18:02:41.088103Z 0 [System] [MY-013602] [Server] Channel mysql_main configured to support TLS. Encrypted connections are now supported for this channel.
2026-08-08T18:02:41.096201Z 0 [Warning] [MY-011810] [Server] Insecure configuration for --pid-file: Location '/var/run/mysqld' in the path is accessible to all OS users. Consider choosing a different directory.
2026-08-08T18:02:41.126004Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Socket: /var/run/mysqld/mysqlx.sock
2026-08-08T18:02:41.126169Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '8.0.46'  socket: '/var/run/mysqld/mysqld.sock'  port: 0  MySQL Community Server - GPL.
2026-08-08 18:02:41+00:00 [Note] [Entrypoint]: Temporary server started.
'/var/lib/mysql/mysql.sock' -> '/var/run/mysqld/mysqld.sock'
Warning: Unable to load '/usr/share/zoneinfo/iso3166.tab' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/leap-seconds.list' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/leapseconds' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/tzdata.zi' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/zone.tab' as time zone. Skipping it.
Warning: Unable to load '/usr/share/zoneinfo/zone1970.tab' as time zone. Skipping it.
2026-08-08 18:02:44+00:00 [Note] [Entrypoint]: Creating database cosmo_db
2026-08-08 18:02:44+00:00 [Note] [Entrypoint]: Creating user cosmo
2026-08-08 18:02:44+00:00 [Note] [Entrypoint]: Giving user cosmo access to schema cosmo_db

2026-08-08 18:02:44+00:00 [Note] [Entrypoint]: Stopping temporary server
2026-08-08T18:02:44.686383Z 15 [System] [MY-013172] [Server] Received SHUTDOWN from user root. Shutting down mysqld (Version: 8.0.46).
2026-08-08T18:02:45.846338Z 0 [System] [MY-010910] [Server] /usr/sbin/mysqld: Shutdown complete (mysqld 8.0.46)  MySQL Community Server - GPL.
2026-08-08 18:02:46+00:00 [Note] [Entrypoint]: Temporary server stopped

2026-08-08 18:02:46+00:00 [Note] [Entrypoint]: MySQL init process done. Ready for start up.

2026-08-08T18:02:47.076794Z 0 [Warning] [MY-011068] [Server] The syntax '--skip-host-cache' is deprecated and will be removed in a future release. Please use SET GLOBAL host_cache_size=0 instead.
2026-08-08T18:02:47.080868Z 0 [System] [MY-010116] [Server] /usr/sbin/mysqld (mysqld 8.0.46) starting as process 1
2026-08-08T18:02:47.105770Z 1 [System] [MY-013576] [InnoDB] InnoDB initialization has started.
2026-08-08T18:02:47.382085Z 1 [System] [MY-013577] [InnoDB] InnoDB initialization has ended.
2026-08-08T18:02:47.705950Z 0 [Warning] [MY-010068] [Server] CA certificate ca.pem is self signed.
2026-08-08T18:02:47.706016Z 0 [System] [MY-013602] [Server] Channel mysql_main configured to support TLS. Encrypted connections are now supported for this channel.
2026-08-08T18:02:47.719822Z 0 [Warning] [MY-011810] [Server] Insecure configuration for --pid-file: Location '/var/run/mysqld' in the path is accessible to all OS users. Consider choosing a different directory.
2026-08-08T18:02:47.752799Z 0 [System] [MY-011323] [Server] X Plugin ready for connections. Bind-address: '::' port: 33060, socket: /var/run/mysqld/mysqlx.sock
2026-08-08T18:02:47.753161Z 0 [System] [MY-010931] [Server] /usr/sbin/mysqld: ready for connections. Version: '8.0.46'  socket: '/var/run/mysqld/mysqld.sock'  port: 3306  MySQL Community Server - GPL.