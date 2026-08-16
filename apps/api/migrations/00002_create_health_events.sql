-- +goose Up

CREATE TABLE health_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL,
    status TEXT NOT NULL,
    response_time_ms INTEGER,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT health_events_service_fk
        FOREIGN KEY (service_id)
        REFERENCES services(id)
        ON DELETE RESTRICT,

    CONSTRAINT health_events_status_valid
        CHECK (status IN ('HEALTHY', 'DEGRADED', 'UNAVAILABLE')),

    CONSTRAINT health_events_response_time_non_negative
        CHECK (
            response_time_ms IS NULL
            OR response_time_ms >= 0
        )
);

CREATE INDEX health_events_service_observed_at_idx
    ON health_events (
        service_id,
        observed_at DESC,
        created_at DESC
    );

-- +goose Down

DROP TABLE health_events;
