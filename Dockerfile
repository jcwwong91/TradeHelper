from iron/base

COPY TradeHelper /
COPY web /web

RUN chmod 400 /web -R

entrypoint /TradeHelper
