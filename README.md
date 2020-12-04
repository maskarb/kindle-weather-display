## kindle-weather-display
* A derivative of
  [https://github.com/mpetroff/kindle-weather-display](https://github.com/mpetroff/kindle-weather-display)

From:
[http://www.mpetroff.net/archives/2012/09/14/kindle-weather-display/](http://www.mpetroff.net/archives/2012/09/14/kindle-weather-display/)

## Refreshed
* Rebuilt using golang for the purposes of learning
* Uses the [ClimaCell Weather API](https://www.climacell.co/weather-api/) (requires a key)

## Dockerfile for Server
* [https://hub.docker.com/r/jtslear/kindle-weather/](https://hub.docker.com/r/jtslear/kindle-weather/) <-- need to update this
* Server requires three environment variables:
  * `CLIMACELL_API_KEY`
  * `LATITUDE`
  * `LONGITUDE`

### Example Run Server
* `docker-compose up --build`

### Example get
* `wget http://localhost:53084/out/output.png`
