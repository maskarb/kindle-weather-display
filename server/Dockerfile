FROM python:3

RUN apt-get update \
  && apt-get install -y librsvg2-bin pngcrush \
  && apt-get clean

ADD weather-script.sh weather-script.py weather-script-preprocess.svg /src/

CMD ["/src/weather-script.sh"]
