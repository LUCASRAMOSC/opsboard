-- +goose Up

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT workspaces_name_not_blank
        CHECK (btrim(name) <> '')
);

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    criticality TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT services_workspace_fk
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT services_name_not_blank
        CHECK (btrim(name) <> ''),

    CONSTRAINT services_type_valid
        CHECK (type IN ('FRONTEND', 'API', 'DATABASE')),

    CONSTRAINT services_criticality_valid
        CHECK (criticality IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),

    CONSTRAINT services_workspace_name_unique
        UNIQUE (workspace_id, name)
);

-- +goose Down

DROP TABLE services;
DROP TABLE workspaces;
