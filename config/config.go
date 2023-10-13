package config

import (
	_ "embed"
)

type Config struct {
	Sites Sites `yaml:"sites" validate:"required"`
	Meta  Meta  `yaml:"meta" validate:"required"`
}

type Sites struct {
	Frontend string `yaml:"frontend" default:"https://stocksim2.narc.live" comment:"Frontend URL" validate:"required"`
	API      string `yaml:"api" default:"https://stocksim2-api.narc.live" comment:"API URL" validate:"required"`
	CDN      string `yaml:"cdn" default:"https://cdn.infinitybots.gg" comment:"CDN URL" validate:"required"`
}

type Meta struct {
	PostgresURL    string   `yaml:"postgres_url" default:"postgresql:///xavage" comment:"Postgres URL" validate:"required"`
	RedisURL       string   `yaml:"redis_url" default:"redis://localhost:6379" comment:"Redis URL" validate:"required"`
	Port           string   `yaml:"port" default:":8081" comment:"Port to run the server on" validate:"required"`
	CDNPath        string   `yaml:"cdn_path" default:"/silverpelt/cdn/ibl" comment:"CDN Path" validate:"required"`
	VulgarList     []string `yaml:"vulgar_list" default:"fuck,suck,shit,kill" validate:"required"`
	UrgentMentions string   `yaml:"urgent_mentions" default:"<@&1061643797315993701>" comment:"Urgent mentions" validate:"required"`
}
