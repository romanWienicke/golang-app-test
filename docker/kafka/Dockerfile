FROM confluentinc/cp-kafka:latest

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
USER root
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
USER appuser

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]