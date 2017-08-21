dp-dataset-api
==================
A ONS API used to navigate datasets which are published.

#### Postgres
* Run ```brew install postgres```
* Run ```brew services start postgres```
* Run ```createuser dp -d -w```
* Run ```createdb --owner dp Datasets```
* Run ```psql -U dp Datasets -f scripts/InitDatabase.sql```

### Configuration

| Environment variable       | Default                                   | Description
| -------------------------- | ----------------------------------------- | -----------
| BIND_ADDR                  | :22000                                    | The host and port to bind to
| POSTGRES_DATASETS_URL      | user=dp dbname=Datasets sslmode=disable   | The URL address to connect to a postgres instance of the database
| SECRET_KEY                 | FD0108EA-825D-411C-9B1D-41EF7727F465      | A secret key used authentication

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

### License

Copyright © 2016-2017, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details
