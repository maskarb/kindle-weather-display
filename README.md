## kindle-weather-display
* A derivative of
  [https://github.com/mpetroff/kindle-weather-display](https://github.com/mpetroff/kindle-weather-display)

From:
[http://www.mpetroff.net/archives/2012/09/14/kindle-weather-display/](http://www.mpetroff.net/archives/2012/09/14/kindle-weather-display/)

## Refreshed
* Rebuilt using golang for the purposes of learning
* Uses the [Darksky API](https://darksky.net/dev/docs) (requires a key)

## Dockerfile for Server
* [https://hub.docker.com/r/jtslear/kindle-weather/](https://hub.docker.com/r/jtslear/kindle-weather/)
* Server requires two environment variables:
  * `DARKSKY_API_KEY`
  * `GPS_COORDINATES` - a comma separated string example: `GPS_COORDINATES='37.8267,-122.4233'`
* The shell script will output the file at `/var/lib/www/weather-script-output.png`
  * Do with this as you please

### Example Run Server
* `docker run -e DARKSKY_API_KEY='thisIsADarSkyApiKey' -e GPS_COORDINATES='37.8267,-122.4233' -v /var/lib/www:/var/lib/www jtslear/kindle-weather`
