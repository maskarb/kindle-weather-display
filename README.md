## kindle-weather-display
* A derivative of
  [https://github.com/mpetroff/kindle-weather-display](https://github.com/mpetroff/kindle-weather-display)

From:
[http://www.mpetroff.net/archives/2012/09/14/kindle-weather-display/](http://www.mpetroff.net/archives/2012/09/14/kindle-weather-display/)

## Refreshed
* Rebuilt using golang for the purposes of learning
* Uses the [ClimaCell Weather API](https://www.climacell.co/weather-api/) (requires a key)

## Dockerfile for Server
* [https://hub.docker.com/r/maskarb/kindle-weather-display/](https://hub.docker.com/r/maskarb/kindle-weather-display/)
* The following environment variables should be set:
  #### Required:
  * `CLIMACELL_API_KEY`
  #### Optional:
  * `LATITUDE` (default is 35.780361)
  * `LONGITUDE` (default is -78.639111)
  * `TIMEZONE` (default is UTC)
  * `CRON_SCHEDULE` (default is `*/5 * * * *`)
* a `.env.example` is included. Copy the example to a `.env` file and update the variables.

### Example Run Server
```
docker run -p 53084:53084 --env-file .env maskarb/kindle-weather-display:kindle-server
```

### Example get
* `wget http://localhost:53084/out/output.png`
