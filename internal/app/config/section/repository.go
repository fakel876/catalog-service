package section

import "time"

type (
	Repository struct {
		Postgres RepositoryPostgres
	}
	RepositoryPostgres struct {
		Address      string        `required:"true"`
		Username     string        `required:"true"`
		Password     string        `required:"true"`
		Name         string        `required:"true"`
		ReadTimeout  time.Duration `default:"5s" split_words:"true"`
		WriteTimeout time.Duration `default:"5s" split_words:"true"`
	}
)
