package coredns_rqlite

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const (
	defaultTtl                = 360
	defaultMaxLifeTime        = 1 * time.Minute
	defaultMaxOpenConnections = 10
	defaultMaxIdleConnections = 10
	defaultZoneUpdateTime     = 10 * time.Minute
)

func init() {
	caddy.RegisterPlugin("rqlite", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	r, err := rqliteParse(c)
	if err != nil {
		return plugin.Error("rqlite", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		return r
	})

	return nil
}

func rqliteParse(c *caddy.Controller) (*CoreDNSRqlite, error) {
	rqlite := CoreDNSRqlite{
		TablePrefix: "coredns_",
		Ttl:         300,
	}
	var err error

	c.Next()
	if c.NextBlock() {
		for {
			switch c.Val() {
			case "dsn":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				rqlite.Dsn = c.Val()
			case "table_prefix":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				rqlite.TablePrefix = c.Val()
			case "max_lifetime":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				var val time.Duration
				val, err = time.ParseDuration(c.Val())
				if err != nil {
					val = defaultMaxLifeTime
				}
				rqlite.MaxLifetime = val
			case "max_open_connections":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultMaxOpenConnections
				}
				rqlite.MaxOpenConnections = val
			case "max_idle_connections":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultMaxIdleConnections
				}
				rqlite.MaxIdleConnections = val
			case "zone_update_interval":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				var val time.Duration
				val, err = time.ParseDuration(c.Val())
				if err != nil {
					val = defaultZoneUpdateTime
				}
				rqlite.zoneUpdateTime = val
			case "ttl":
				if !c.NextArg() {
					return &CoreDNSRqlite{}, c.ArgErr()
				}
				var val int
				val, err = strconv.Atoi(c.Val())
				if err != nil {
					val = defaultTtl
				}
				rqlite.Ttl = uint32(val)
			default:
				if c.Val() != "}" {
					return &CoreDNSRqlite{}, c.Errf("unknown property '%s'", c.Val())
				}
			}

			if !c.Next() {
				break
			}
		}

	}

	db, err := rqlite.db()
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rqlite.tableName = rqlite.TablePrefix + "records"

	return &rqlite, nil
}

func (handler *CoreDNSRqlite) db() (*sql.DB, error) {
	db, err := sql.Open("rqlite", os.ExpandEnv(handler.Dsn))
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(handler.MaxLifetime)
	db.SetMaxOpenConns(handler.MaxOpenConnections)
	db.SetMaxIdleConns(handler.MaxIdleConnections)

	return db, nil
}
