FROM otel/opentelemetry-collector-opampsupervisor:latest as supervisor

# Get OpenTelemetry collector
FROM otel/opentelemetry-collector-contrib:latest AS otel

# Final stage
FROM alpine:latest

# Copy from builder - use the correct source path
COPY --from=supervisor /usr/local/bin/opampsupervisor ./opampsupervisor
COPY supervisor.yaml .
COPY collector.yaml .

# Copy from OpenTelemetry
COPY --from=otel /otelcol-contrib /otelcol-contrib

EXPOSE 4317 4318

ENTRYPOINT ["./opampsupervisor", "--config", "supervisor.yaml"]