package config

type Config struct {
	DatabaseURL string
}

var Default = Config{
	DatabaseURL: "postgres://user:password@localhost:5439/sagas_dev?sslmode=disable",
}
