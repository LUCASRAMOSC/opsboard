-- +goose Up

CREATE TABLE business_journeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name TEXT NOT NULL,
    criticality TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT business_journeys_workspace_fk
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT business_journeys_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT business_journeys_criticality_valid
        CHECK (
            criticality IN (
                'LOW',
                'MEDIUM',
                'HIGH',
                'CRITICAL'
            )
        ),

    CONSTRAINT business_journeys_workspace_name_unique
        UNIQUE (workspace_id, name)
);

CREATE TABLE business_journey_services (
    business_journey_id UUID NOT NULL,
    service_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT business_journey_services_pk
        PRIMARY KEY (
            business_journey_id,
            service_id
        ),

    CONSTRAINT business_journey_services_journey_fk
        FOREIGN KEY (business_journey_id)
        REFERENCES business_journeys(id)
        ON DELETE CASCADE,

    CONSTRAINT business_journey_services_service_fk
        FOREIGN KEY (service_id)
        REFERENCES services(id)
        ON DELETE CASCADE
);

CREATE INDEX business_journey_services_service_id_idx
    ON business_journey_services (service_id);

-- +goose Down

DROP TABLE business_journey_services;
DROP TABLE business_journeys;
